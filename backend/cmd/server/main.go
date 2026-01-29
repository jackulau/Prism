package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jacklau/prism/internal/agent"
	"github.com/jacklau/prism/internal/api/routes"
	"github.com/jacklau/prism/internal/api/sse"
	"github.com/jacklau/prism/internal/api/websocket"
	"github.com/jacklau/prism/internal/cloudprovider"
	"github.com/jacklau/prism/internal/config"
	"github.com/jacklau/prism/internal/database"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations"
	"github.com/jacklau/prism/internal/integrations/discord"
	"github.com/jacklau/prism/internal/integrations/posthog"
	"github.com/jacklau/prism/internal/integrations/slack"
	"github.com/jacklau/prism/internal/llm"
	"github.com/jacklau/prism/internal/llm/anthropic"
	"github.com/jacklau/prism/internal/llm/deepseek"
	"github.com/jacklau/prism/internal/llm/google"
	"github.com/jacklau/prism/internal/llm/lmstudio"
	"github.com/jacklau/prism/internal/llm/ollama"
	"github.com/jacklau/prism/internal/llm/openai"
	"github.com/jacklau/prism/internal/mcp"
	"github.com/jacklau/prism/internal/sandbox"
	"github.com/jacklau/prism/internal/security"
	"github.com/jacklau/prism/internal/services/coderunner"
	"github.com/jacklau/prism/internal/tools"
	"github.com/jacklau/prism/internal/tools/builtin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.NewSQLite(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Initialize security services
	encryptionService, err := security.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to create encryption service: %v", err)
	}

	jwtService := security.NewJWTService(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	conversationRepo := repository.NewConversationRepository(db.DB)
	messageRepo := repository.NewMessageRepository(db.DB)
	webhookRepo := repository.NewWebhookRepository(db.DB, encryptionService)
	providerKeyRepo := repository.NewProviderKeyRepository(db.DB)
	integrationRepo := repository.NewIntegrationRepository(db.DB, encryptionService)
	fileHistoryRepo := repository.NewFileHistoryRepository(db.DB)
	workspaceRepo := repository.NewWorkspaceRepository(db.DB)
	todoRepo := repository.NewTodoRepository(db.DB)
	dataConfigRepo := repository.NewDataConfigRepository(db.DB, encryptionService)

	// Initialize code runner for GitHub webhook automation
	var codeRunner *coderunner.Runner
	if cfg.CodeRunnerEnabled {
		codeRunner = coderunner.NewRunner(&coderunner.Config{
			DockerEnabled: cfg.CodeRunnerDockerMode,
			MemoryLimit:   cfg.CodeRunnerMemoryLimit,
			CPULimit:      cfg.CodeRunnerCPULimit,
			Timeout:       cfg.CodeRunnerTimeout,
		})
		log.Println("Code runner initialized")
	}

	// Initialize sandbox service for terminal/build functionality
	sandboxService, err := sandbox.NewService(cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize sandbox service: %v", err)
	} else {
		log.Println("Sandbox service initialized")
		// Attach workspace repository for workspace persistence
		sandboxService.SetWorkspaceRepository(workspaceRepo)
	}

	// Initialize tool registry with built-in tools
	toolRegistry := tools.NewRegistry()
	if sandboxService != nil {
		toolConfig := builtin.Config{
			FileHistoryRepo:   fileHistoryRepo,
			TodoRepo:          todoRepo,
			UserRepo:          userRepo,
			EncryptionService: encryptionService,
		}
		if err := builtin.RegisterAll(toolRegistry, sandboxService, codeRunner, db.DB, toolConfig); err != nil {
			log.Printf("Warning: Failed to register built-in tools: %v", err)
		} else {
			log.Println("Built-in tools registered")
		}
	}

	// Initialize LLM manager
	llmManager := llm.NewManager()

	// Register LLM providers
	// Ollama (local LLM - no API key needed)
	ollamaClient := ollama.NewClient(cfg.OllamaHost)
	llmManager.RegisterProvider(ollamaClient)

	// LM Studio (local LLM - no API key needed)
	lmstudioClient := lmstudio.NewClient(cfg.LMStudioHost)
	llmManager.RegisterProvider(lmstudioClient)

	// OpenAI (API key set via UI)
	openaiClient := openai.NewClient("")
	llmManager.RegisterProvider(openaiClient)

	// Anthropic (API key set via UI)
	anthropicClient := anthropic.NewClient("")
	llmManager.RegisterProvider(anthropicClient)

	// Google AI (API key set via UI)
	googleClient := google.NewClient("")
	llmManager.RegisterProvider(googleClient)

	// DeepSeek (API key set via UI)
	deepseekClient := deepseek.NewClient("")
	llmManager.RegisterProvider(deepseekClient)

	log.Printf("Registered %d LLM providers", len(llmManager.ListProviders()))

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize SSE service for Server-Sent Events streaming
	sseService := sse.NewService()
	log.Println("SSE service initialized")

	// Initialize integrations manager
	integrationManager := integrations.NewManager()

	// Register Discord integration
	discordClient := discord.NewClient(&discord.Config{
		WebhookURL: cfg.DiscordWebhookURL,
		BotToken:   cfg.DiscordBotToken,
		Enabled:    cfg.DiscordEnabled,
	})
	integrationManager.RegisterNotification(discordClient)

	// Register Slack integration
	slackClient := slack.NewClient(&slack.Config{
		WebhookURL: cfg.SlackWebhookURL,
		BotToken:   cfg.SlackBotToken,
		ChannelID:  cfg.SlackChannelID,
		Enabled:    cfg.SlackEnabled,
	})
	integrationManager.RegisterNotification(slackClient)

	// Register PostHog integration
	posthogClient := posthog.NewClient(&posthog.Config{
		APIKey:        cfg.PostHogAPIKey,
		Endpoint:      cfg.PostHogEndpoint,
		Enabled:       cfg.PostHogEnabled,
		BatchSize:     cfg.PostHogBatchSize,
		FlushInterval: cfg.PostHogFlushInterval,
	})
	integrationManager.RegisterAnalytics(posthogClient)

	// Initialize agent manager for parallel agent execution
	agentManager := agent.NewManager(llmManager, agent.DefaultManagerConfig())
	agentManager.SetAgentRepository(agentRepo)
	agentManager.Start()
	log.Println("Agent manager started")

	// Initialize MCP components
	mcpServer := mcp.NewServer(toolRegistry)
	mcpClient := mcp.NewClient()
	mcpRepo := mcp.NewRepository(db.DB)

	// Load all enabled HTTP MCP connections from database
	if err := mcpRepo.LoadAllEnabled(mcpClient); err != nil {
		log.Printf("Warning: Failed to load HTTP MCP connections: %v", err)
	}

	// Initialize stdio MCP client for local MCP servers
	stdioMCPClient := mcp.NewStdioClient()
	stdioMCPRepo := mcp.NewStdioRepository(db.DB)

	// Load all enabled stdio MCP servers from database
	if err := stdioMCPRepo.LoadAllEnabled(stdioMCPClient); err != nil {
		log.Printf("Warning: Failed to load stdio MCP servers: %v", err)
	}
	log.Println("MCP server and clients initialized")

	// Initialize cloud provider manager
	cloudProviderManager := cloudprovider.NewManager()
	// Note: Cloud providers will be registered when their implementations are available
	// (e.g., claude-cloud, openai-assistants)
	log.Println("Cloud provider manager initialized")

	// Setup routes
	deps := &routes.Dependencies{
		Config:             cfg,
		JWTService:         jwtService,
		EncryptionService:  encryptionService,
		UserRepo:           userRepo,
		SessionRepo:        sessionRepo,
		ConversationRepo:   conversationRepo,
		MessageRepo:        messageRepo,
		WebhookRepo:        webhookRepo,
		ProviderKeyRepo:    providerKeyRepo,
		IntegrationRepo:    integrationRepo,
		FileHistoryRepo:    fileHistoryRepo,
		DataConfigRepo:     dataConfigRepo,
		LLMManager:         llmManager,
		WSHub:              wsHub,
		SSEService:         sseService,
		IntegrationManager: integrationManager,
		AgentManager:       agentManager,
		CodeRunner:         codeRunner,
		SandboxService:     sandboxService,
		ToolRegistry:       toolRegistry,
		MCPServer:              mcpServer,
		MCPClient:              mcpClient,
		MCPRepository:          mcpRepo,
		StdioMCPClient:         stdioMCPClient,
		StdioMCPRepository:     stdioMCPRepo,
		CloudProviderManager:   cloudProviderManager,
	}

	app := routes.Setup(deps)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Println("Gracefully shutting down...")

		// Stop agent manager
		agentManager.Stop()
		log.Println("Agent manager stopped")

		// Stop all stdio MCP servers
		stdioMCPClient.StopAll()
		log.Println("Stdio MCP servers stopped")

		// Stop tunnel server if running
		if tunnelServer != nil {
			if err := tunnelServer.Stop(); err != nil {
				log.Printf("Error stopping tunnel server: %v", err)
			}
			log.Println("Tunnel server stopped")
		}

		// Close integrations manager
		if err := integrationManager.Close(); err != nil {
			log.Printf("Error closing integrations: %v", err)
		}

		if err := app.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Start server
	addr := cfg.Host + ":" + cfg.Port
	log.Printf("Starting Prism server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
