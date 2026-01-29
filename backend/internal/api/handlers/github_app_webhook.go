package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations/github"
)

// GitHubAppWebhookHandler handles GitHub App webhook events
type GitHubAppWebhookHandler struct {
	app              *github.GitHubApp
	installationRepo *repository.GitHubInstallationRepo
	webhookSecret    string
}

// NewGitHubAppWebhookHandler creates a new webhook handler for GitHub App events
func NewGitHubAppWebhookHandler(
	app *github.GitHubApp,
	installationRepo *repository.GitHubInstallationRepo,
	webhookSecret string,
) *GitHubAppWebhookHandler {
	return &GitHubAppWebhookHandler{
		app:              app,
		installationRepo: installationRepo,
		webhookSecret:    webhookSecret,
	}
}

// HandleWebhook handles incoming GitHub App webhook events
func (h *GitHubAppWebhookHandler) HandleWebhook(c *fiber.Ctx) error {
	eventType := c.Get("X-GitHub-Event")
	signature := c.Get("X-Hub-Signature-256")
	deliveryID := c.Get("X-GitHub-Delivery")

	if eventType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing X-GitHub-Event header",
		})
	}

	body := c.Body()

	// Verify signature if webhook secret is configured
	if h.webhookSecret != "" {
		if err := github.VerifySignature(body, signature, h.webhookSecret); err != nil {
			log.Printf("Webhook signature verification failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid signature",
			})
		}
	}

	log.Printf("Received GitHub App webhook: event=%s, delivery=%s", eventType, deliveryID)

	// Parse the event
	event, err := github.ParseEvent(eventType, body)
	if err != nil {
		log.Printf("Failed to parse webhook event: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse event",
		})
	}

	// Handle the event based on type
	switch eventType {
	case "installation":
		if err := h.handleInstallationEvent(event); err != nil {
			log.Printf("Failed to handle installation event: %v", err)
			// Return 200 to prevent GitHub from retrying
		}
	case "installation_repositories":
		if err := h.handleInstallationRepositoriesEvent(event); err != nil {
			log.Printf("Failed to handle installation_repositories event: %v", err)
		}
	case "ping":
		log.Printf("Received ping from GitHub")
	default:
		log.Printf("Unhandled webhook event type: %s", eventType)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"event":   eventType,
	})
}

// handleInstallationEvent processes installation webhook events
func (h *GitHubAppWebhookHandler) handleInstallationEvent(event interface{}) error {
	installEvent, ok := event.(*github.InstallationEvent)
	if !ok {
		log.Printf("Expected InstallationEvent but got: %T", event)
		return nil
	}

	log.Printf("Processing installation event: action=%s, installation_id=%d, account=%s",
		installEvent.Action, installEvent.Installation.ID, installEvent.Installation.Account.Login)

	switch installEvent.Action {
	case "created":
		return h.handleInstalled(installEvent)
	case "deleted":
		return h.handleUninstalled(installEvent)
	case "suspend":
		return h.handleSuspended(installEvent)
	case "unsuspend":
		return h.handleUnsuspended(installEvent)
	default:
		log.Printf("Unknown installation action: %s", installEvent.Action)
	}

	return nil
}

// handleInstalled handles when the App is installed
func (h *GitHubAppWebhookHandler) handleInstalled(event *github.InstallationEvent) error {
	inst := event.Installation

	// Convert permissions to map
	permsMap := make(map[string]string)
	if inst.Permissions != nil {
		if inst.Permissions.Contents != "" {
			permsMap["contents"] = inst.Permissions.Contents
		}
		if inst.Permissions.Issues != "" {
			permsMap["issues"] = inst.Permissions.Issues
		}
		if inst.Permissions.Metadata != "" {
			permsMap["metadata"] = inst.Permissions.Metadata
		}
		if inst.Permissions.PullRequests != "" {
			permsMap["pull_requests"] = inst.Permissions.PullRequests
		}
	}

	dbInstallation := &repository.GitHubInstallation{
		InstallationID:      inst.ID,
		AccountID:           inst.Account.ID,
		AccountLogin:        inst.Account.Login,
		AccountType:         inst.Account.Type,
		AccountAvatarURL:    inst.Account.AvatarURL,
		AppID:               inst.AppID,
		TargetType:          inst.TargetType,
		Permissions:         repository.EncodePermissions(permsMap),
		Events:              repository.EncodeEvents(inst.Events),
		RepositorySelection: inst.RepositorySelection,
	}

	// Check if installation already exists
	existing, err := h.installationRepo.GetByInstallationID(inst.ID)
	if err != nil {
		log.Printf("Failed to check existing installation: %v", err)
		return err
	}

	if existing != nil {
		dbInstallation.ID = existing.ID
		if err := h.installationRepo.Update(dbInstallation); err != nil {
			log.Printf("Failed to update installation: %v", err)
			return err
		}
	} else {
		if err := h.installationRepo.Create(dbInstallation); err != nil {
			log.Printf("Failed to create installation: %v", err)
			return err
		}
	}

	// Save initial repositories if provided
	if len(event.Repositories) > 0 {
		repos := make([]*repository.GitHubInstallationRepository, len(event.Repositories))
		for i, repo := range event.Repositories {
			repos[i] = &repository.GitHubInstallationRepository{
				InstallationID: inst.ID,
				RepositoryID:   repo.ID,
				FullName:       repo.FullName,
				Name:           repo.Name,
				Private:        repo.Private,
				HTMLURL:        repo.HTMLURL,
				Description:    repo.Description,
			}
		}

		if err := h.installationRepo.SetRepositories(inst.ID, repos); err != nil {
			log.Printf("Failed to save repositories: %v", err)
		}
	}

	log.Printf("Installation created: %s (%s)", inst.Account.Login, inst.Account.Type)
	return nil
}

// handleUninstalled handles when the App is uninstalled
func (h *GitHubAppWebhookHandler) handleUninstalled(event *github.InstallationEvent) error {
	inst := event.Installation

	if err := h.installationRepo.Delete(inst.ID); err != nil {
		log.Printf("Failed to delete installation: %v", err)
		return err
	}

	// Invalidate any cached tokens
	if h.app != nil {
		h.app.InvalidateToken(inst.ID)
	}

	log.Printf("Installation deleted: %s", inst.Account.Login)
	return nil
}

// handleSuspended handles when the App installation is suspended
func (h *GitHubAppWebhookHandler) handleSuspended(event *github.InstallationEvent) error {
	inst := event.Installation

	if err := h.installationRepo.Suspend(inst.ID); err != nil {
		log.Printf("Failed to suspend installation: %v", err)
		return err
	}

	// Invalidate any cached tokens
	if h.app != nil {
		h.app.InvalidateToken(inst.ID)
	}

	log.Printf("Installation suspended: %s", inst.Account.Login)
	return nil
}

// handleUnsuspended handles when the App installation is unsuspended
func (h *GitHubAppWebhookHandler) handleUnsuspended(event *github.InstallationEvent) error {
	inst := event.Installation

	if err := h.installationRepo.Unsuspend(inst.ID); err != nil {
		log.Printf("Failed to unsuspend installation: %v", err)
		return err
	}

	log.Printf("Installation unsuspended: %s", inst.Account.Login)
	return nil
}

// handleInstallationRepositoriesEvent processes repository access change events
func (h *GitHubAppWebhookHandler) handleInstallationRepositoriesEvent(event interface{}) error {
	repoEvent, ok := event.(*github.InstallationRepositoriesEvent)
	if !ok {
		log.Printf("Expected InstallationRepositoriesEvent but got: %T", event)
		return nil
	}

	installationID := repoEvent.Installation.ID
	log.Printf("Processing installation_repositories event: action=%s, installation_id=%d",
		repoEvent.Action, installationID)

	switch repoEvent.Action {
	case "added":
		return h.handleRepositoriesAdded(repoEvent)
	case "removed":
		return h.handleRepositoriesRemoved(repoEvent)
	default:
		log.Printf("Unknown installation_repositories action: %s", repoEvent.Action)
	}

	return nil
}

// handleRepositoriesAdded handles when repositories are added to the App
func (h *GitHubAppWebhookHandler) handleRepositoriesAdded(event *github.InstallationRepositoriesEvent) error {
	installationID := event.Installation.ID

	for _, repo := range event.RepositoriesAdded {
		dbRepo := &repository.GitHubInstallationRepository{
			InstallationID: installationID,
			RepositoryID:   repo.ID,
			FullName:       repo.FullName,
			Name:           repo.Name,
			Private:        repo.Private,
			HTMLURL:        repo.HTMLURL,
			Description:    repo.Description,
		}

		if err := h.installationRepo.AddRepository(dbRepo); err != nil {
			log.Printf("Failed to add repository %s: %v", repo.FullName, err)
		} else {
			log.Printf("Repository added: %s", repo.FullName)
		}
	}

	return nil
}

// handleRepositoriesRemoved handles when repositories are removed from the App
func (h *GitHubAppWebhookHandler) handleRepositoriesRemoved(event *github.InstallationRepositoriesEvent) error {
	installationID := event.Installation.ID

	for _, repo := range event.RepositoriesRemoved {
		if err := h.installationRepo.RemoveRepository(installationID, repo.ID); err != nil {
			log.Printf("Failed to remove repository %s: %v", repo.FullName, err)
		} else {
			log.Printf("Repository removed: %s", repo.FullName)
		}
	}

	return nil
}

