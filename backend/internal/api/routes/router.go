package routes

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/api/handlers"
	"github.com/jacklau/prism/internal/api/middleware"
	ws "github.com/jacklau/prism/internal/api/websocket"
	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/mcp"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/security"
	"github.com/jacklau/prism/internal/services/coderunner"
	"github.com/jacklau/prism/internal/tools"
)

// Dependencies holds all the dependencies for the router
type Dependencies struct {
	Config             *config.Config
	JWTService         *security.JWTService
	EncryptionService  *security.EncryptionService
	UserRepo           *repository.UserRepository
	SessionRepo        *repository.SessionRepository
	ConversationRepo   *repository.ConversationRepository
	MessageRepo        *repository.MessageRepository
	WebhookRepo        *repository.WebhookRepository
	ProviderKeyRepo    *repository.ProviderKeyRepository
	IntegrationRepo    *repository.IntegrationRepository
	FileHistoryRepo    *repository.FileHistoryRepository
	LLMManager         *llm.Manager
	WSHub              *ws.Hub
	IntegrationManager *integrations.Manager
	AgentManager       *agent.Manager
	CodeRunner         *coderunner.Runner
	SandboxService     *sandbox.Service
	ToolRegistry       *tools.Registry
	MCPServer          *mcp.Server
	MCPClient          *mcp.Client
	MCPRepository      *mcp.Repository
	StdioMCPClient     *mcp.StdioClient
	StdioMCPRepository *mcp.StdioRepository
}

// Setup sets up the Fiber app with all routes
func Setup(deps *Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middleware.SecurityHeaders())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     deps.Config.CORSAllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// API v1
	v1 := app.Group("/api/v1")

	// Auth routes (no auth required)
	authHandler := handlers.NewAuthHandler(deps.UserRepo, deps.SessionRepo, deps.JWTService)
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)

	// Guest login route (if enabled)
	if deps.Config.GuestModeEnabled {
		auth.Post("/guest", authHandler.GuestLogin)
		v1.Get("/guest-mode", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"enabled": true})
		})
		log.Println("Guest mode enabled - users can access without registration")
	} else {
		v1.Get("/guest-mode", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"enabled": false})
		})
	}

	// Auth routes (auth required)
	authProtected := auth.Group("", middleware.AuthMiddleware(deps.JWTService))
	authProtected.Post("/logout", authHandler.Logout)
	authProtected.Get("/me", authHandler.Me)

	// Chat routes (auth required)
	chatHandler := handlers.NewChatHandler(deps.ConversationRepo, deps.MessageRepo)
	conversations := v1.Group("/conversations", middleware.AuthMiddleware(deps.JWTService))
	conversations.Get("/", chatHandler.ListConversations)
	conversations.Get("/search", chatHandler.SearchConversations)
	conversations.Post("/", chatHandler.CreateConversation)
	conversations.Get("/:id", chatHandler.GetConversation)
	conversations.Patch("/:id", chatHandler.UpdateConversation)
	conversations.Delete("/:id", chatHandler.DeleteConversation)
	conversations.Get("/:id/messages", chatHandler.GetMessages)

	// WebSocket route
	v1.Use("/ws", func(c *fiber.Ctx) error {
		// Check for WebSocket upgrade
		if websocket.IsWebSocketUpgrade(c) {
			// Try to get token from Sec-WebSocket-Protocol header first (more secure)
			// Format: "auth, <token>" - we use "auth" as a marker and token as second protocol
			protocols := c.Get("Sec-WebSocket-Protocol")
			var token string

			if protocols != "" {
				// Parse protocols - expecting "auth, <token>"
				parts := strings.Split(protocols, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "auth" && p != "" {
						token = p
						break
					}
				}
			}

			// Fallback to query parameter for backwards compatibility
			if token == "" {
				token = c.Query("token")
			}

			if token == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "missing token",
				})
			}

			claims, err := deps.JWTService.ValidateAccessToken(token)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "invalid token",
				})
			}

			c.Locals("userID", claims.UserID)
			c.Locals("email", claims.Email)
			// Store that we should respond with the auth protocol
			c.Locals("wsProtocol", "auth")
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	v1.Get("/ws", websocket.New(func(c *websocket.Conn) {
		userID := c.Locals("userID").(string)

		client := ws.NewClient(deps.WSHub, c, userID, func(client *ws.Client, msg *ws.IncomingMessage) {
			// Handle incoming messages
			handleWebSocketMessage(deps, client, msg)
		})

		deps.WSHub.Register(client)

		// Start read/write pumps
		go client.WritePump()
		client.ReadPump()
	}, websocket.Config{
		// Accept "auth" subprotocol - required for browser to keep connection open
		// when client sends Sec-WebSocket-Protocol: auth, <token>
		Subprotocols: []string{"auth"},
	}))

	// Provider routes (auth required)
	providers := v1.Group("/providers", middleware.AuthMiddleware(deps.JWTService))
	providers.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"providers": deps.LLMManager.ListProviders(),
		})
	})

	// Provider key management routes
	if deps.ProviderKeyRepo != nil {
		providerHandler := handlers.NewProviderHandler(deps.ProviderKeyRepo, deps.EncryptionService, deps.LLMManager)
		providers.Post("/:provider/key", providerHandler.SetKey)
		providers.Delete("/:provider/key", providerHandler.DeleteKey)
		providers.Post("/:provider/validate", providerHandler.ValidateKey)
		providers.Get("/:provider/key/status", providerHandler.GetKeyStatus)
		providers.Get("/keys", providerHandler.ListKeys)
	}

	// Preview/Sandbox routes (auth required)
	if deps.SandboxService != nil {
		previewHandler := handlers.NewPreviewHandler(deps.SandboxService)
		sandbox := v1.Group("/sandbox", middleware.AuthMiddleware(deps.JWTService))
		sandbox.Get("/files", previewHandler.ListFiles)
		sandbox.Get("/files/*", previewHandler.GetFile)
		sandbox.Post("/files", previewHandler.WriteFile)
		sandbox.Delete("/files/*", previewHandler.DeleteFile)
		sandbox.Get("/builds/:id", previewHandler.GetBuild)
		sandbox.Post("/builds/:id/stop", previewHandler.StopBuild)

		// Preview server (serves static files from sandbox)
		app.Get("/preview/:userID/*", previewHandler.ServePreview)

		// Workspace management routes (auth required)
		workspaceHandler := handlers.NewWorkspaceHandler(deps.SandboxService)
		workspace := v1.Group("/workspace", middleware.AuthMiddleware(deps.JWTService))
		workspace.Get("/directory", workspaceHandler.GetDirectory)
		workspace.Post("/directory", workspaceHandler.SetDirectory)
		workspace.Get("/browse", workspaceHandler.BrowseDirectories)
		workspace.Post("/pick-folder", workspaceHandler.OpenFolderPicker)
		workspace.Get("/recent", workspaceHandler.ListRecentWorkspaces)
		workspace.Post("/:id/current", workspaceHandler.SetCurrentWorkspace)
		workspace.Delete("/:id", workspaceHandler.RemoveWorkspace)
	}

	// GitHub webhook routes
	if deps.WebhookRepo != nil && deps.CodeRunner != nil {
		githubHandler := handlers.NewGitHubHandler(
			deps.WebhookRepo,
			deps.CodeRunner,
			deps.Config.GitHubWebhookSecret,
			deps.IntegrationManager,
		)

		// Public webhook endpoint (no auth - verified by signature)
		v1.Post("/github/webhook", githubHandler.HandleWebhook)

		// Webhook configuration routes (auth required)
		github := v1.Group("/github", middleware.AuthMiddleware(deps.JWTService))
		github.Get("/webhooks", githubHandler.GetWebhookConfigs)
		github.Post("/webhooks", githubHandler.CreateWebhookConfig)
		github.Get("/webhooks/:id", githubHandler.GetWebhookConfig)
		github.Patch("/webhooks/:id", githubHandler.UpdateWebhookConfig)
		github.Delete("/webhooks/:id", githubHandler.DeleteWebhookConfig)
		github.Post("/webhooks/:id/test", githubHandler.TestWebhook)
		github.Get("/webhooks/:id/deliveries", githubHandler.GetWebhookDeliveries)

		// Code execution endpoint (auth required)
		github.Post("/run", githubHandler.RunCode)
	}

	// OAuth routes
	if deps.Config.GitHubClientID != "" {
		oauthHandler := handlers.NewOAuthHandler(deps.UserRepo, deps.EncryptionService, deps.Config)

		// Public callback (OAuth redirect - no auth, user identified by state)
		v1.Get("/oauth/github/callback", oauthHandler.GitHubCallback)

		// Protected OAuth routes
		oauth := v1.Group("/oauth", middleware.AuthMiddleware(deps.JWTService))
		oauth.Get("/github/authorize", oauthHandler.GitHubAuthorize)

		// GitHub account management (protected)
		githubAccount := v1.Group("/github", middleware.AuthMiddleware(deps.JWTService))
		githubAccount.Get("/status", oauthHandler.GitHubStatus)
		githubAccount.Delete("/disconnect", oauthHandler.DisconnectGitHub)
		githubAccount.Get("/repos", oauthHandler.ListGitHubRepos)

		// GitHub clone (requires sandbox service)
		if deps.SandboxService != nil {
			workspaceHandler := handlers.NewWorkspaceHandler(deps.SandboxService)
			githubAccount.Post("/clone", workspaceHandler.CloneGitHubRepo)
		}
	}

	// MCP routes
	if deps.MCPServer != nil {
		// Register MCP server routes (exposes tools to external clients)
		deps.MCPServer.RegisterRoutes(v1)
	}

	if deps.MCPClient != nil && deps.MCPRepository != nil {
		// Register MCP client routes (connect to external HTTP MCP servers)
		mcpHandler := mcp.NewHandler(deps.MCPClient, deps.MCPRepository)
		mcpProtected := v1.Group("", middleware.AuthMiddleware(deps.JWTService))
		mcpHandler.RegisterRoutes(mcpProtected)
	}

	if deps.StdioMCPClient != nil && deps.StdioMCPRepository != nil {
		// Register stdio MCP routes (connect to local MCP servers via stdin/stdout)
		stdioHandler := mcp.NewStdioHandler(deps.StdioMCPClient, deps.StdioMCPRepository)
		stdioProtected := v1.Group("", middleware.AuthMiddleware(deps.JWTService))
		stdioHandler.RegisterRoutes(stdioProtected)
	}

	// Integrations routes (for Settings page)
	integrationsRoute := v1.Group("/integrations", middleware.AuthMiddleware(deps.JWTService))
	if deps.IntegrationRepo != nil {
		integrationHandler := handlers.NewIntegrationHandler(deps.IntegrationRepo)
		integrationsRoute.Get("/status", integrationHandler.GetStatus)
		integrationsRoute.Post("/discord", integrationHandler.SetDiscord)
		integrationsRoute.Delete("/discord", integrationHandler.DeleteDiscord)
		integrationsRoute.Post("/slack", integrationHandler.SetSlack)
		integrationsRoute.Delete("/slack", integrationHandler.DeleteSlack)
		integrationsRoute.Post("/posthog", integrationHandler.SetPostHog)
		integrationsRoute.Delete("/posthog", integrationHandler.DeletePostHog)
	} else {
		// Fallback to config-based status if no repo
		integrationsRoute.Get("/status", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"discord": fiber.Map{
					"enabled":   deps.Config.DiscordEnabled,
					"connected": deps.Config.DiscordWebhookURL != "",
				},
				"slack": fiber.Map{
					"enabled":   deps.Config.SlackEnabled,
					"connected": deps.Config.SlackWebhookURL != "",
				},
				"posthog": fiber.Map{
					"enabled":   deps.Config.PostHogEnabled,
					"connected": deps.Config.PostHogAPIKey != "",
				},
			})
		})
	}

	return app
}

// handleWebSocketMessage handles incoming WebSocket messages
func handleWebSocketMessage(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	switch msg.Type {
	case ws.TypeChatMessage:
		// Track message sent event
		if deps.IntegrationManager != nil {
			deps.IntegrationManager.TrackMessageSent(client.UserID, msg.ConversationID, "")
		}

		// Handle chat message with LLM streaming
		handleChatMessage(deps, client, msg)

	case ws.TypeChatStop:
		// Track chat stopped event
		if deps.IntegrationManager != nil {
			deps.IntegrationManager.Track(&integrations.Event{
				Type:           integrations.EventChatStopped,
				UserID:         client.UserID,
				ConversationID: msg.ConversationID,
			})
		}

		// Handle stopping generation
		handleChatStop(deps, client, msg)

	case ws.TypeToolConfirm:
		// Track tool approval/rejection
		if deps.IntegrationManager != nil {
			eventType := integrations.EventToolApproved
			if !msg.Approved {
				eventType = integrations.EventToolRejected
			}
			deps.IntegrationManager.TrackAndNotify(&integrations.Event{
				Type:           eventType,
				UserID:         client.UserID,
				ConversationID: msg.ConversationID,
				Data: map[string]interface{}{
					"execution_id": msg.ExecutionID,
					"approved":     msg.Approved,
				},
			})
		}

		// Handle tool confirmation
		handleToolConfirm(deps, client, msg)

	// Agent message handlers
	case ws.TypeAgentRun:
		handleAgentRun(deps, client, msg)

	case ws.TypeAgentRunParallel:
		handleAgentRunParallel(deps, client, msg)

	case ws.TypeAgentStop:
		handleAgentStop(deps, client, msg)

	case ws.TypeAgentStatus:
		handleAgentStatus(deps, client, msg)

	case ws.TypeAgentList:
		handleAgentList(deps, client, msg)

	// Swarm/Multi-agent message handlers
	case ws.TypeSwarmRun:
		handleSwarmRun(deps, client, msg)

	case ws.TypeSwarmStop:
		handleSwarmStop(deps, client, msg)

	case ws.TypeSwarmStatus:
		handleSwarmStatus(deps, client, msg)

	case ws.TypeSwarmList:
		handleSwarmList(deps, client, msg)

	// Preview/Sandbox message handlers
	case ws.TypeBuildStart:
		handleBuildStart(deps, client, msg)

	case ws.TypeBuildStop:
		handleBuildStop(deps, client, msg)

	case ws.TypeFileRequest:
		handleFileRequest(deps, client, msg)

	case ws.TypeFileHistoryRequest:
		handleFileHistoryRequest(deps, client, msg)

	default:
		client.SendMessage(ws.NewError("unknown_type", "unknown message type: "+msg.Type))

		// Track error event
		if deps.IntegrationManager != nil {
			deps.IntegrationManager.TrackError(client.UserID, msg.ConversationID, "unknown_type", "unknown message type: "+msg.Type)
		}
	}
}

// handleAgentRun handles a single agent run request
func handleAgentRun(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.AgentConfig == nil {
		client.SendMessage(ws.NewError("invalid_request", "agent_config is required"))
		return
	}

	if msg.Content == "" {
		client.SendMessage(ws.NewError("invalid_request", "content (prompt) is required"))
		return
	}

	// Create agent config
	agentConfig := agent.AgentConfig{
		Name:         msg.AgentConfig.Name,
		Provider:     msg.AgentConfig.Provider,
		Model:        msg.AgentConfig.Model,
		SystemPrompt: msg.AgentConfig.SystemPrompt,
		Temperature:  msg.AgentConfig.Temperature,
		MaxTokens:    msg.AgentConfig.MaxTokens,
	}

	// Create task
	task := agent.NewTask(msg.Content,
		agent.WithContext(msg.Context),
		agent.WithPriority(agent.TaskPriority(msg.Priority)),
	)

	// Run the agent
	execution, err := deps.AgentManager.RunTask(context.Background(), task, agentConfig)
	if err != nil {
		client.SendMessage(ws.NewError("agent_error", err.Error()))
		return
	}

	// Subscribe to events and forward them to the client
	go forwardAgentEvents(deps, client, execution)

	log.Printf("Agent started: id=%s, task=%s", execution.Agents[0].ID, task.ID)
}

// handleAgentRunParallel handles a parallel agent run request
func handleAgentRunParallel(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.AgentConfig == nil {
		client.SendMessage(ws.NewError("invalid_request", "agent_config is required"))
		return
	}

	if len(msg.Tasks) == 0 {
		client.SendMessage(ws.NewError("invalid_request", "tasks array is required"))
		return
	}

	// Create agent config
	agentConfig := agent.AgentConfig{
		Name:         msg.AgentConfig.Name,
		Provider:     msg.AgentConfig.Provider,
		Model:        msg.AgentConfig.Model,
		SystemPrompt: msg.AgentConfig.SystemPrompt,
		Temperature:  msg.AgentConfig.Temperature,
		MaxTokens:    msg.AgentConfig.MaxTokens,
	}

	// Create tasks
	tasks := make([]*agent.Task, len(msg.Tasks))
	for i, t := range msg.Tasks {
		tasks[i] = agent.NewTask(t.Prompt,
			agent.WithTaskID(t.ID),
			agent.WithContext(t.Context),
			agent.WithMetadata(t.Metadata),
		)
	}

	// Run agents in parallel
	execution, err := deps.AgentManager.RunParallel(context.Background(), tasks, agentConfig)
	if err != nil {
		client.SendMessage(ws.NewError("agent_error", err.Error()))
		return
	}

	// Forward events and batch progress to client
	go forwardBatchEvents(deps, client, execution)

	log.Printf("Parallel agent execution started: id=%s, tasks=%d", execution.ID, len(tasks))
}

// handleAgentStop handles an agent stop request
func handleAgentStop(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.ExecutionID == "" && msg.AgentID == "" {
		client.SendMessage(ws.NewError("invalid_request", "execution_id or agent_id is required"))
		return
	}

	var err error
	if msg.ExecutionID != "" {
		err = deps.AgentManager.CancelExecution(msg.ExecutionID)
	}

	if err != nil {
		client.SendMessage(ws.NewError("agent_error", err.Error()))
		return
	}

	client.SendMessage(&ws.OutgoingMessage{
		Type:        ws.TypeAgentCancelled,
		ExecutionID: msg.ExecutionID,
		AgentID:     msg.AgentID,
		Status:      "cancelled",
	})
}

// handleAgentStatus handles an agent status request
func handleAgentStatus(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.ExecutionID == "" {
		client.SendMessage(ws.NewError("invalid_request", "execution_id is required"))
		return
	}

	execution, err := deps.AgentManager.GetExecution(msg.ExecutionID)
	if err != nil {
		client.SendMessage(ws.NewError("agent_error", err.Error()))
		return
	}

	// Convert agents to AgentInfo
	agents := make([]ws.AgentInfo, len(execution.Agents))
	for i, a := range execution.Agents {
		agents[i] = ws.AgentInfo{
			ID:        a.ID,
			Name:      a.Config.Name,
			Status:    string(a.GetStatus()),
			Provider:  a.Config.Provider,
			Model:     a.Config.Model,
			CreatedAt: a.CreatedAt.UnixMilli(),
		}
		if a.StartedAt != nil {
			ts := a.StartedAt.UnixMilli()
			agents[i].StartedAt = &ts
		}
		if a.CompletedAt != nil {
			ts := a.CompletedAt.UnixMilli()
			agents[i].CompletedAt = &ts
		}
	}

	client.SendMessage(&ws.OutgoingMessage{
		Type:        ws.TypeAgentStatus,
		ExecutionID: execution.ID,
		Status:      string(execution.GetStatus()),
		Agents:      agents,
	})
}

// handleAgentList handles an agent list request
func handleAgentList(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	executions := deps.AgentManager.ListExecutions()

	agents := make([]ws.AgentInfo, 0)
	for _, exec := range executions {
		for _, a := range exec.Agents {
			info := ws.AgentInfo{
				ID:        a.ID,
				Name:      a.Config.Name,
				Status:    string(a.GetStatus()),
				Provider:  a.Config.Provider,
				Model:     a.Config.Model,
				CreatedAt: a.CreatedAt.UnixMilli(),
			}
			if a.StartedAt != nil {
				ts := a.StartedAt.UnixMilli()
				info.StartedAt = &ts
			}
			if a.CompletedAt != nil {
				ts := a.CompletedAt.UnixMilli()
				info.CompletedAt = &ts
			}
			agents = append(agents, info)
		}
	}

	client.SendMessage(ws.NewAgentList(agents))
}

// forwardAgentEvents forwards agent events to the WebSocket client
func forwardAgentEvents(deps *Dependencies, client *ws.Client, execution *agent.Execution) {
	if len(execution.Agents) == 0 {
		return
	}

	agentInstance := execution.Agents[0]
	taskID := ""
	if len(execution.Tasks) > 0 {
		taskID = execution.Tasks[0].ID
	}

	// Send started notification
	client.SendMessage(ws.NewAgentStarted(agentInstance.ID, taskID))

	// Listen to agent events
	for event := range agentInstance.Events() {
		switch event.Type {
		case agent.AgentEventStreamChunk:
			if delta, ok := event.Data["delta"].(string); ok {
				client.SendMessage(ws.NewAgentStreamChunk(agentInstance.ID, taskID, delta))
			}
		case agent.AgentEventToolCall:
			toolName, _ := event.Data["name"].(string)
			params := event.Data["parameters"]
			client.SendMessage(ws.NewAgentToolCall(agentInstance.ID, taskID, toolName, params))
		case agent.AgentEventCompleted:
			output, _ := event.Data["output"].(string)
			// Wait for result to get duration
			select {
			case result := <-agentInstance.Results():
				client.SendMessage(ws.NewAgentCompleted(agentInstance.ID, taskID, output, result.Duration.Milliseconds()))
			case <-time.After(5 * time.Second):
				client.SendMessage(ws.NewAgentCompleted(agentInstance.ID, taskID, output, 0))
			}
			return
		case agent.AgentEventFailed:
			errMsg, _ := event.Data["error"].(string)
			client.SendMessage(ws.NewAgentFailed(agentInstance.ID, taskID, errMsg))
			return
		case agent.AgentEventCancelled:
			client.SendMessage(ws.NewAgentCancelled(agentInstance.ID, taskID))
			return
		}
	}
}

// forwardBatchEvents forwards batch execution events to the WebSocket client
func forwardBatchEvents(deps *Dependencies, client *ws.Client, execution *agent.Execution) {
	startTime := time.Now()
	totalTasks := len(execution.Tasks)

	// Track progress
	completedTasks := 0
	failedTasks := 0

	// Subscribe to agent manager events
	eventChan := deps.AgentManager.Subscribe()
	defer deps.AgentManager.Unsubscribe(eventChan)

	// Send initial progress
	client.SendMessage(ws.NewAgentBatchProgress(execution.ID, &ws.BatchProgressInfo{
		TotalTasks:     totalTasks,
		CompletedTasks: 0,
		FailedTasks:    0,
		RunningTasks:   totalTasks,
		PendingTasks:   0,
	}))

	// Listen for events from agents in this execution
	agentIDs := make(map[string]bool)
	for _, a := range execution.Agents {
		agentIDs[a.ID] = true
	}

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return
			}

			// Only process events for our agents
			if !agentIDs[event.AgentID] {
				continue
			}

			switch event.Type {
			case agent.AgentEventStreamChunk:
				// Forward stream chunks for individual agents
				if delta, ok := event.Data["delta"].(string); ok {
					// Find task ID for this agent
					var taskID string
					for i, a := range execution.Agents {
						if a.ID == event.AgentID && i < len(execution.Tasks) {
							taskID = execution.Tasks[i].ID
							break
						}
					}
					client.SendMessage(ws.NewAgentStreamChunk(event.AgentID, taskID, delta))
				}

			case agent.AgentEventCompleted:
				completedTasks++
				client.SendMessage(ws.NewAgentBatchProgress(execution.ID, &ws.BatchProgressInfo{
					TotalTasks:     totalTasks,
					CompletedTasks: completedTasks,
					FailedTasks:    failedTasks,
					RunningTasks:   totalTasks - completedTasks - failedTasks,
					PendingTasks:   0,
				}))

			case agent.AgentEventFailed:
				failedTasks++
				completedTasks++
				client.SendMessage(ws.NewAgentBatchProgress(execution.ID, &ws.BatchProgressInfo{
					TotalTasks:     totalTasks,
					CompletedTasks: completedTasks,
					FailedTasks:    failedTasks,
					RunningTasks:   totalTasks - completedTasks - failedTasks,
					PendingTasks:   0,
				}))
			}

			// Check if all tasks are done
			if completedTasks >= totalTasks {
				// Get final results
				results := execution.GetResults()
				resultInfos := make([]ws.AgentResultInfo, len(results))
				for i, r := range results {
					resultInfos[i] = ws.AgentResultInfo{
						AgentID:     r.AgentID,
						TaskID:      r.TaskID,
						Success:     r.Success,
						Output:      r.Output,
						Error:       r.Error,
						Duration:    r.Duration.Milliseconds(),
						CompletedAt: r.CompletedAt.UnixMilli(),
					}
				}

				client.SendMessage(ws.NewAgentBatchCompleted(
					execution.ID,
					resultInfos,
					time.Since(startTime).Milliseconds(),
				))
				return
			}

		case <-time.After(10 * time.Minute):
			// Timeout for batch execution
			client.SendMessage(ws.NewError("batch_timeout", "batch execution timed out"))
			return
		}
	}
}

// ==================== Swarm/Multi-Agent Handlers ====================

// handleSwarmRun handles a swarm run request
func handleSwarmRun(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.Content == "" {
		client.SendMessage(ws.NewError("invalid_request", "content (task) is required"))
		return
	}

	// Build agent role configs
	var agentConfigs []agent.AgentRoleConfig

	if msg.SwarmConfig != nil && len(msg.SwarmConfig.AgentRoles) > 0 {
		// Use explicit swarm config
		for _, rc := range msg.SwarmConfig.AgentRoles {
			roleConfig := agent.AgentRoleConfig{
				Role:         agent.AgentRole(rc.Role),
				Count:        rc.Count,
				SystemPrompt: rc.SystemPrompt,
			}
			if rc.Config != nil {
				roleConfig.Config = agent.AgentConfig{
					Provider:    rc.Config.Provider,
					Model:       rc.Config.Model,
					Temperature: rc.Config.Temperature,
					MaxTokens:   rc.Config.MaxTokens,
				}
			}
			agentConfigs = append(agentConfigs, roleConfig)
		}
	} else if len(msg.AgentRoles) > 0 {
		// Use roles from message
		for _, rc := range msg.AgentRoles {
			roleConfig := agent.AgentRoleConfig{
				Role:         agent.AgentRole(rc.Role),
				Count:        rc.Count,
				SystemPrompt: rc.SystemPrompt,
			}
			if rc.Config != nil {
				roleConfig.Config = agent.AgentConfig{
					Provider:    rc.Config.Provider,
					Model:       rc.Config.Model,
					Temperature: rc.Config.Temperature,
					MaxTokens:   rc.Config.MaxTokens,
				}
			}
			agentConfigs = append(agentConfigs, roleConfig)
		}
	} else if msg.AgentConfig != nil {
		// Create default parallel agents from base config
		agentConfigs = []agent.AgentRoleConfig{
			{Role: agent.RoleCoder, Count: 1},
			{Role: agent.RoleReviewer, Count: 1},
			{Role: agent.RoleTester, Count: 1},
		}
	} else {
		client.SendMessage(ws.NewError("invalid_request", "swarm_config, agent_roles, or agent_config is required"))
		return
	}

	// Get base config
	baseConfig := agent.AgentConfig{
		Provider: "openai",
		Model:    "gpt-4",
	}
	if msg.AgentConfig != nil {
		baseConfig = agent.AgentConfig{
			Provider:    msg.AgentConfig.Provider,
			Model:       msg.AgentConfig.Model,
			Temperature: msg.AgentConfig.Temperature,
			MaxTokens:   msg.AgentConfig.MaxTokens,
		}
	}

	// Determine strategy
	strategy := agent.StrategyParallel
	if msg.Strategy != "" {
		strategy = agent.SwarmStrategy(msg.Strategy)
	} else if msg.SwarmConfig != nil && msg.SwarmConfig.Strategy != "" {
		strategy = agent.SwarmStrategy(msg.SwarmConfig.Strategy)
	}

	// Run the swarm
	swarm, err := deps.AgentManager.RunMultiAgent(context.Background(), msg.Content, strategy, agentConfigs, baseConfig)
	if err != nil {
		client.SendMessage(ws.NewError("swarm_error", err.Error()))
		return
	}

	// Forward swarm events to client
	go forwardSwarmEvents(deps, client, swarm)

	log.Printf("Swarm started: id=%s, agents=%d, strategy=%s", swarm.ID, len(swarm.Agents), strategy)
}

// handleSwarmStop handles a swarm stop request
func handleSwarmStop(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.SwarmID == "" {
		client.SendMessage(ws.NewError("invalid_request", "swarm_id is required"))
		return
	}

	if err := deps.AgentManager.CancelSwarm(msg.SwarmID); err != nil {
		client.SendMessage(ws.NewError("swarm_error", err.Error()))
		return
	}

	client.SendMessage(ws.NewSwarmCancelled(msg.SwarmID))
}

// handleSwarmStatus handles a swarm status request
func handleSwarmStatus(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	if msg.SwarmID == "" {
		client.SendMessage(ws.NewError("invalid_request", "swarm_id is required"))
		return
	}

	swarm, err := deps.AgentManager.GetSwarm(msg.SwarmID)
	if err != nil {
		client.SendMessage(ws.NewError("swarm_error", err.Error()))
		return
	}

	// Convert to websocket types
	agents := make([]ws.SwarmAgentInfo, len(swarm.Agents))
	running, completed, failed := 0, 0, 0

	for i, sa := range swarm.Agents {
		agents[i] = ws.SwarmAgentInfo{
			ID:     sa.ID,
			Role:   string(sa.Role),
			Status: string(sa.Status),
			Input:  sa.Input,
			Output: sa.Output,
		}
		if sa.StartedAt != nil {
			ts := sa.StartedAt.UnixMilli()
			agents[i].StartedAt = &ts
		}
		if sa.CompletedAt != nil {
			ts := sa.CompletedAt.UnixMilli()
			agents[i].CompletedAt = &ts
		}

		switch sa.Status {
		case agent.AgentStatusRunning:
			running++
		case agent.AgentStatusCompleted:
			completed++
		case agent.AgentStatusFailed:
			failed++
		}
	}

	progress := &ws.SwarmProgressInfo{
		TotalAgents:     len(swarm.Agents),
		RunningAgents:   running,
		CompletedAgents: completed,
		FailedAgents:    failed,
		Phase:           string(swarm.Status),
	}

	client.SendMessage(ws.NewSwarmStatus(swarm.ID, string(swarm.Status), agents, progress))
}

// handleSwarmList handles a swarm list request
func handleSwarmList(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.AgentManager == nil {
		client.SendMessage(ws.NewError("agent_unavailable", "agent manager not available"))
		return
	}

	swarms := deps.AgentManager.ListSwarms()

	// Convert to info list
	infos := make([]ws.SwarmAgentInfo, 0)
	for _, s := range swarms {
		for _, sa := range s.Agents {
			info := ws.SwarmAgentInfo{
				ID:     sa.ID,
				Role:   string(sa.Role),
				Status: string(sa.Status),
			}
			infos = append(infos, info)
		}
	}

	client.SendMessage(ws.NewSwarmList(infos))
}

// forwardSwarmEvents forwards swarm events to the WebSocket client
func forwardSwarmEvents(deps *Dependencies, client *ws.Client, swarm *agent.Swarm) {
	startTime := time.Now()

	// Build initial agent info
	agents := make([]ws.SwarmAgentInfo, len(swarm.Agents))
	for i, sa := range swarm.Agents {
		agents[i] = ws.SwarmAgentInfo{
			ID:     sa.ID,
			Role:   string(sa.Role),
			Status: string(sa.Status),
		}
	}

	// Send swarm started
	client.SendMessage(ws.NewSwarmStarted(swarm.ID, agents))

	// Track progress
	totalAgents := len(swarm.Agents)
	completedAgents := 0
	failedAgents := 0

	// Listen for events
	for event := range swarm.Events() {
		switch event.Type {
		case agent.SwarmEventAgentStarted:
			input, _ := event.Data["input"].(string)
			client.SendMessage(ws.NewSwarmAgentStarted(swarm.ID, event.AgentID, string(event.Role), input))

		case agent.SwarmEventAgentOutput:
			if delta, ok := event.Data["delta"].(string); ok {
				client.SendMessage(ws.NewSwarmAgentOutput(swarm.ID, event.AgentID, string(event.Role), delta))
			}

		case agent.SwarmEventAgentCompleted:
			completedAgents++
			output, _ := event.Data["output"].(string)
			duration, _ := event.Data["duration"].(int64)
			client.SendMessage(ws.NewSwarmAgentCompleted(swarm.ID, event.AgentID, string(event.Role), output, duration))

			// Send progress update
			client.SendMessage(ws.NewSwarmProgress(swarm.ID, &ws.SwarmProgressInfo{
				TotalAgents:     totalAgents,
				RunningAgents:   totalAgents - completedAgents - failedAgents,
				CompletedAgents: completedAgents,
				FailedAgents:    failedAgents,
				Phase:           "running",
			}))

		case agent.SwarmEventAgentFailed:
			failedAgents++
			completedAgents++
			errMsg, _ := event.Data["error"].(string)
			client.SendMessage(ws.NewSwarmAgentFailed(swarm.ID, event.AgentID, string(event.Role), errMsg))

		case agent.SwarmEventSynthesizing:
			client.SendMessage(ws.NewSwarmSynthesizing(swarm.ID))
			client.SendMessage(ws.NewSwarmProgress(swarm.ID, &ws.SwarmProgressInfo{
				TotalAgents:     totalAgents,
				RunningAgents:   0,
				CompletedAgents: completedAgents,
				FailedAgents:    failedAgents,
				Phase:           "synthesizing",
			}))

		case agent.SwarmEventCompleted:
			// Get final agents info
			finalAgents := make([]ws.SwarmAgentInfo, len(swarm.Agents))
			for i, sa := range swarm.Agents {
				finalAgents[i] = ws.SwarmAgentInfo{
					ID:     sa.ID,
					Role:   string(sa.Role),
					Status: string(sa.Status),
					Output: sa.Output,
				}
				if sa.StartedAt != nil {
					ts := sa.StartedAt.UnixMilli()
					finalAgents[i].StartedAt = &ts
				}
				if sa.CompletedAt != nil {
					ts := sa.CompletedAt.UnixMilli()
					finalAgents[i].CompletedAt = &ts
				}
			}

			client.SendMessage(ws.NewSwarmCompleted(
				swarm.ID,
				swarm.FinalOutput,
				finalAgents,
				time.Since(startTime).Milliseconds(),
			))
			return

		case agent.SwarmEventFailed:
			errMsg, _ := event.Data["error"].(string)
			client.SendMessage(ws.NewSwarmFailed(swarm.ID, errMsg))
			return

		case agent.SwarmEventCancelled:
			client.SendMessage(ws.NewSwarmCancelled(swarm.ID))
			return
		}
	}
}

// handleBuildStart handles a build start request via WebSocket
func handleBuildStart(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.SandboxService == nil {
		client.SendMessage(ws.NewError("sandbox_unavailable", "sandbox service not available"))
		return
	}

	// Get build command from params
	command := "npm"
	args := []string{"run", "dev"}

	if msg.Params != nil {
		if cmd, ok := msg.Params["command"].(string); ok && cmd != "" {
			command = cmd
		}
		if a, ok := msg.Params["args"].([]interface{}); ok {
			args = make([]string, len(a))
			for i, arg := range a {
				if s, ok := arg.(string); ok {
					args[i] = s
				}
			}
		}
	}

	// Start the build
	var buildID string
	build, err := deps.SandboxService.StartBuild(client.UserID, command, args, func(line sandbox.OutputLine) {
		// Forward build output to client via WebSocket
		client.SendMessage(ws.NewBuildOutput(buildID, line.Content, line.Stream))
	})
	if build != nil {
		buildID = build.ID
	}
	if err != nil {
		client.SendMessage(ws.NewError("build_error", err.Error()))
		return
	}

	// Send build started notification
	client.SendMessage(ws.NewBuildStarted(build.ID))

	// Monitor build completion in background
	go func() {
		// Wait for build to complete by polling status
		for {
			time.Sleep(500 * time.Millisecond)
			b, err := deps.SandboxService.GetBuild(build.ID)
			if err != nil {
				return
			}
			if b.Status == sandbox.BuildStatusSuccess || b.Status == sandbox.BuildStatusFailed || b.Status == sandbox.BuildStatusCancelled {
				var duration int64
				if b.EndTime != nil {
					duration = b.EndTime.Sub(b.StartTime).Milliseconds()
				}
				previewURL := deps.SandboxService.GetPreviewServer(client.UserID)
				client.SendMessage(ws.NewBuildCompleted(b.ID, b.Status == sandbox.BuildStatusSuccess, previewURL, duration))

				// Also send files updated message
				files, err := deps.SandboxService.ListFiles(client.UserID)
				if err == nil {
					wsFiles := make([]ws.FileInfo, len(files))
					for i, f := range files {
						wsFiles[i] = convertFileInfo(f)
					}
					client.SendMessage(ws.NewFilesUpdated(wsFiles))
				}
				return
			}
		}
	}()
}

// handleBuildStop handles a build stop request via WebSocket
func handleBuildStop(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.SandboxService == nil {
		client.SendMessage(ws.NewError("sandbox_unavailable", "sandbox service not available"))
		return
	}

	buildID := ""
	if msg.Params != nil {
		if id, ok := msg.Params["build_id"].(string); ok {
			buildID = id
		}
	}

	if buildID == "" {
		client.SendMessage(ws.NewError("invalid_request", "build_id is required"))
		return
	}

	if err := deps.SandboxService.StopBuild(buildID); err != nil {
		client.SendMessage(ws.NewError("build_error", err.Error()))
		return
	}

	client.SendMessage(&ws.OutgoingMessage{
		Type:    ws.TypeBuildCompleted,
		BuildID: buildID,
		Status:  "cancelled",
		Success: false,
	})
}

// handleFileRequest handles a file content request via WebSocket
func handleFileRequest(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.SandboxService == nil {
		client.SendMessage(ws.NewError("sandbox_unavailable", "sandbox service not available"))
		return
	}

	filePath := ""
	if msg.Params != nil {
		if p, ok := msg.Params["path"].(string); ok {
			filePath = p
		}
	}

	if filePath == "" {
		client.SendMessage(ws.NewError("invalid_request", "path is required"))
		return
	}

	content, err := deps.SandboxService.GetFileContent(client.UserID, filePath)
	if err != nil {
		client.SendMessage(ws.NewError("file_error", err.Error()))
		return
	}

	client.SendMessage(ws.NewFileContent(filePath, content))
}

// convertFileInfo converts sandbox.FileInfo to ws.FileInfo
func convertFileInfo(f sandbox.FileInfo) ws.FileInfo {
	wsFile := ws.FileInfo{
		Name:        f.Name,
		Path:        f.Path,
		IsDirectory: f.IsDirectory,
		Size:        f.Size,
		Modified:    f.Modified,
	}
	if len(f.Children) > 0 {
		wsFile.Children = make([]ws.FileInfo, len(f.Children))
		for i, child := range f.Children {
			wsFile.Children[i] = convertFileInfo(child)
		}
	}
	return wsFile
}

// handleFileHistoryRequest handles file history list and content requests via WebSocket
func handleFileHistoryRequest(deps *Dependencies, client *ws.Client, msg *ws.IncomingMessage) {
	if deps.FileHistoryRepo == nil {
		client.SendMessage(ws.NewError("history_unavailable", "file history not available"))
		return
	}

	// Get action from params (list or get)
	action := ""
	if msg.Params != nil {
		if a, ok := msg.Params["action"].(string); ok {
			action = a
		}
	}

	switch action {
	case "get":
		// Get content of a specific history entry
		historyID := ""
		if msg.Params != nil {
			if id, ok := msg.Params["history_id"].(string); ok {
				historyID = id
			}
		}
		if historyID == "" {
			client.SendMessage(ws.NewError("invalid_request", "history_id is required"))
			return
		}

		entry, err := deps.FileHistoryRepo.GetByID(historyID)
		if err != nil {
			client.SendMessage(ws.NewError("history_error", err.Error()))
			return
		}
		if entry == nil || entry.UserID != client.UserID {
			client.SendMessage(ws.NewError("history_error", "history entry not found"))
			return
		}

		client.SendMessage(ws.NewFileHistoryContent(
			entry.ID,
			entry.FilePath,
			entry.Content,
			entry.Operation,
			entry.CreatedAt.Format("2006-01-02 15:04:05"),
		))

	default:
		// Default: list history entries
		filePath := ""
		limit := 20
		if msg.Params != nil {
			if p, ok := msg.Params["path"].(string); ok {
				filePath = p
			}
			if l, ok := msg.Params["limit"].(float64); ok {
				limit = int(l)
			}
		}

		var history []*repository.FileHistory
		var err error

		if filePath != "" {
			history, err = deps.FileHistoryRepo.ListByFilePath(client.UserID, filePath, limit)
		} else {
			history, err = deps.FileHistoryRepo.ListByUserID(client.UserID, limit, 0)
		}

		if err != nil {
			client.SendMessage(ws.NewError("history_error", err.Error()))
			return
		}

		entries := make([]ws.FileHistoryEntry, len(history))
		for i, h := range history {
			entries[i] = ws.FileHistoryEntry{
				ID:        h.ID,
				FilePath:  h.FilePath,
				Operation: h.Operation,
				Size:      len(h.Content),
				CreatedAt: h.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}

		client.SendMessage(ws.NewFileHistoryList(filePath, entries))
	}
}

// errorHandler handles errors globally
func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
