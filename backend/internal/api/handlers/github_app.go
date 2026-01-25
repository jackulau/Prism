package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/database/repository"
	"github.com/jacklau/prism/internal/integrations/github"
)

// GitHubAppHandler handles GitHub App-related endpoints
type GitHubAppHandler struct {
	app              *github.GitHubApp
	installationRepo *repository.GitHubInstallationRepo
}

// NewGitHubAppHandler creates a new GitHub App handler
func NewGitHubAppHandler(app *github.GitHubApp, installationRepo *repository.GitHubInstallationRepo) *GitHubAppHandler {
	return &GitHubAppHandler{
		app:              app,
		installationRepo: installationRepo,
	}
}

// GetStatus returns the GitHub App configuration status
func (h *GitHubAppHandler) GetStatus(c *fiber.Ctx) error {
	configured := h.app != nil && h.app.IsConfigured()

	return c.JSON(fiber.Map{
		"configured": configured,
		"app_id":     h.app.AppID(),
	})
}

// ListInstallations lists all GitHub App installations
func (h *GitHubAppHandler) ListInstallations(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	// Fetch installations from database
	installations, err := h.installationRepo.List()
	if err != nil {
		log.Printf("Failed to list installations from database: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list installations",
		})
	}

	// Convert to response format
	response := make([]fiber.Map, len(installations))
	for i, inst := range installations {
		response[i] = fiber.Map{
			"id":                   inst.ID,
			"installation_id":     inst.InstallationID,
			"account_login":       inst.AccountLogin,
			"account_type":        inst.AccountType,
			"account_avatar_url":  inst.AccountAvatarURL,
			"repository_selection": inst.RepositorySelection,
			"permissions":         repository.DecodePermissions(inst.Permissions),
			"events":              repository.DecodeEvents(inst.Events),
			"suspended":           inst.SuspendedAt != nil,
			"created_at":          inst.CreatedAt,
		}
	}

	return c.JSON(fiber.Map{
		"installations": response,
	})
}

// GetInstallation retrieves a specific installation by ID
func (h *GitHubAppHandler) GetInstallation(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	installationID, err := c.ParamsInt("installationID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid installation ID",
		})
	}

	installation, err := h.installationRepo.GetByInstallationID(int64(installationID))
	if err != nil {
		log.Printf("Failed to get installation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get installation",
		})
	}

	if installation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "installation not found",
		})
	}

	return c.JSON(fiber.Map{
		"id":                   installation.ID,
		"installation_id":     installation.InstallationID,
		"account_login":       installation.AccountLogin,
		"account_type":        installation.AccountType,
		"account_avatar_url":  installation.AccountAvatarURL,
		"repository_selection": installation.RepositorySelection,
		"permissions":         repository.DecodePermissions(installation.Permissions),
		"events":              repository.DecodeEvents(installation.Events),
		"suspended":           installation.SuspendedAt != nil,
		"created_at":          installation.CreatedAt,
	})
}

// ListInstallationRepos lists repositories for a specific installation
func (h *GitHubAppHandler) ListInstallationRepos(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	installationID, err := c.ParamsInt("installationID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid installation ID",
		})
	}

	// First check if installation exists
	installation, err := h.installationRepo.GetByInstallationID(int64(installationID))
	if err != nil {
		log.Printf("Failed to get installation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get installation",
		})
	}

	if installation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "installation not found",
		})
	}

	// Get repos from database
	repos, err := h.installationRepo.ListRepositories(int64(installationID))
	if err != nil {
		log.Printf("Failed to list repositories: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list repositories",
		})
	}

	// Convert to response format
	response := make([]fiber.Map, len(repos))
	for i, repo := range repos {
		response[i] = fiber.Map{
			"id":          repo.ID,
			"repository_id": repo.RepositoryID,
			"full_name":   repo.FullName,
			"name":        repo.Name,
			"private":     repo.Private,
			"html_url":    repo.HTMLURL,
			"description": repo.Description,
		}
	}

	return c.JSON(fiber.Map{
		"repositories": response,
		"total_count":  len(repos),
	})
}

// RefreshInstallationRepos fetches repositories from GitHub API and updates the database
func (h *GitHubAppHandler) RefreshInstallationRepos(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	installationID, err := c.ParamsInt("installationID")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid installation ID",
		})
	}

	// First check if installation exists
	installation, err := h.installationRepo.GetByInstallationID(int64(installationID))
	if err != nil {
		log.Printf("Failed to get installation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get installation",
		})
	}

	if installation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "installation not found",
		})
	}

	// Fetch repos from GitHub API
	client := github.NewAppClient(h.app, int64(installationID))
	ghRepos, err := client.ListRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories from GitHub: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to fetch repositories from GitHub",
		})
	}

	// Convert and save to database
	dbRepos := make([]*repository.GitHubInstallationRepository, len(ghRepos))
	for i, repo := range ghRepos {
		dbRepos[i] = &repository.GitHubInstallationRepository{
			InstallationID: int64(installationID),
			RepositoryID:   repo.ID,
			FullName:       repo.FullName,
			Name:           repo.Name,
			Private:        repo.Private,
			HTMLURL:        repo.HTMLURL,
			Description:    repo.Description,
		}
	}

	if err := h.installationRepo.SetRepositories(int64(installationID), dbRepos); err != nil {
		log.Printf("Failed to save repositories: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save repositories",
		})
	}

	// Convert to response format
	response := make([]fiber.Map, len(dbRepos))
	for i, repo := range dbRepos {
		response[i] = fiber.Map{
			"id":          repo.ID,
			"repository_id": repo.RepositoryID,
			"full_name":   repo.FullName,
			"name":        repo.Name,
			"private":     repo.Private,
			"html_url":    repo.HTMLURL,
			"description": repo.Description,
		}
	}

	return c.JSON(fiber.Map{
		"repositories": response,
		"total_count":  len(response),
		"refreshed":    true,
	})
}

// SetupRequest represents the request body for GitHub App setup
type SetupRequest struct {
	InstallationID int64  `json:"installation_id"`
	SetupAction    string `json:"setup_action"` // "install" or "update"
}

// HandleSetup handles the GitHub App installation callback
// This is called when a user installs the GitHub App
func (h *GitHubAppHandler) HandleSetup(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	// Parse query parameters (sent by GitHub after installation)
	installationID := c.QueryInt("installation_id", 0)
	setupAction := c.Query("setup_action", "")

	if installationID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing installation_id",
		})
	}

	log.Printf("GitHub App setup: installation_id=%d, action=%s", installationID, setupAction)

	// Fetch installation details from GitHub API
	installation, err := h.app.GetInstallation(int64(installationID))
	if err != nil {
		log.Printf("Failed to get installation from GitHub: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to verify installation with GitHub",
		})
	}

	// Convert permissions to JSON string
	permsMap := make(map[string]string)
	if installation.Permissions != nil {
		if installation.Permissions.Contents != "" {
			permsMap["contents"] = installation.Permissions.Contents
		}
		if installation.Permissions.Issues != "" {
			permsMap["issues"] = installation.Permissions.Issues
		}
		if installation.Permissions.Metadata != "" {
			permsMap["metadata"] = installation.Permissions.Metadata
		}
		if installation.Permissions.PullRequests != "" {
			permsMap["pull_requests"] = installation.Permissions.PullRequests
		}
	}

	// Check if installation already exists
	existing, err := h.installationRepo.GetByInstallationID(int64(installationID))
	if err != nil {
		log.Printf("Failed to check existing installation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process installation",
		})
	}

	dbInstallation := &repository.GitHubInstallation{
		InstallationID:      installation.ID,
		AccountID:           installation.Account.ID,
		AccountLogin:        installation.Account.Login,
		AccountType:         installation.Account.Type,
		AccountAvatarURL:    installation.Account.AvatarURL,
		AppID:               installation.AppID,
		TargetType:          installation.TargetType,
		Permissions:         repository.EncodePermissions(permsMap),
		Events:              repository.EncodeEvents(installation.Events),
		RepositorySelection: installation.RepositorySelection,
	}

	if existing != nil {
		// Update existing installation
		dbInstallation.ID = existing.ID
		if err := h.installationRepo.Update(dbInstallation); err != nil {
			log.Printf("Failed to update installation: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update installation",
			})
		}
	} else {
		// Create new installation
		if err := h.installationRepo.Create(dbInstallation); err != nil {
			log.Printf("Failed to create installation: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save installation",
			})
		}
	}

	// Fetch and save repositories
	client := github.NewAppClient(h.app, int64(installationID))
	ghRepos, err := client.ListRepositories()
	if err != nil {
		log.Printf("Failed to fetch repositories: %v", err)
		// Don't fail the setup, just log the error
	} else {
		dbRepos := make([]*repository.GitHubInstallationRepository, len(ghRepos))
		for i, repo := range ghRepos {
			dbRepos[i] = &repository.GitHubInstallationRepository{
				InstallationID: int64(installationID),
				RepositoryID:   repo.ID,
				FullName:       repo.FullName,
				Name:           repo.Name,
				Private:        repo.Private,
				HTMLURL:        repo.HTMLURL,
				Description:    repo.Description,
			}
		}

		if err := h.installationRepo.SetRepositories(int64(installationID), dbRepos); err != nil {
			log.Printf("Failed to save repositories: %v", err)
		}
	}

	return c.JSON(fiber.Map{
		"success":         true,
		"installation_id": installationID,
		"account":         installation.Account.Login,
		"action":          setupAction,
	})
}

// ListAllRepos lists all repositories across all installations
func (h *GitHubAppHandler) ListAllRepos(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	installations, err := h.installationRepo.List()
	if err != nil {
		log.Printf("Failed to list installations: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list installations",
		})
	}

	var allRepos []fiber.Map
	for _, inst := range installations {
		repos, err := h.installationRepo.ListRepositories(inst.InstallationID)
		if err != nil {
			log.Printf("Failed to list repos for installation %d: %v", inst.InstallationID, err)
			continue
		}

		for _, repo := range repos {
			allRepos = append(allRepos, fiber.Map{
				"id":              repo.ID,
				"repository_id":   repo.RepositoryID,
				"full_name":       repo.FullName,
				"name":            repo.Name,
				"private":         repo.Private,
				"html_url":        repo.HTMLURL,
				"description":     repo.Description,
				"installation_id": inst.InstallationID,
				"account_login":   inst.AccountLogin,
				"account_type":    inst.AccountType,
			})
		}
	}

	return c.JSON(fiber.Map{
		"repositories": allRepos,
		"total_count":  len(allRepos),
	})
}

// SyncAllInstallations syncs all installations with GitHub API
func (h *GitHubAppHandler) SyncAllInstallations(c *fiber.Ctx) error {
	if h.app == nil || !h.app.IsConfigured() {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "GitHub App not configured",
		})
	}

	// Fetch all installations from GitHub API
	ghInstallations, err := h.app.ListInstallations()
	if err != nil {
		log.Printf("Failed to fetch installations from GitHub: %v", err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": "failed to fetch installations from GitHub",
		})
	}

	synced := 0
	for _, inst := range ghInstallations {
		// Convert permissions
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

		existing, err := h.installationRepo.GetByInstallationID(inst.ID)
		if err != nil {
			log.Printf("Failed to check existing installation %d: %v", inst.ID, err)
			continue
		}

		if existing != nil {
			dbInstallation.ID = existing.ID
			if err := h.installationRepo.Update(dbInstallation); err != nil {
				log.Printf("Failed to update installation %d: %v", inst.ID, err)
				continue
			}
		} else {
			if err := h.installationRepo.Create(dbInstallation); err != nil {
				log.Printf("Failed to create installation %d: %v", inst.ID, err)
				continue
			}
		}

		// Sync repositories
		client := github.NewAppClient(h.app, inst.ID)
		ghRepos, err := client.ListRepositories()
		if err != nil {
			log.Printf("Failed to fetch repos for installation %d: %v", inst.ID, err)
			continue
		}

		dbRepos := make([]*repository.GitHubInstallationRepository, len(ghRepos))
		for i, repo := range ghRepos {
			dbRepos[i] = &repository.GitHubInstallationRepository{
				InstallationID: inst.ID,
				RepositoryID:   repo.ID,
				FullName:       repo.FullName,
				Name:           repo.Name,
				Private:        repo.Private,
				HTMLURL:        repo.HTMLURL,
				Description:    repo.Description,
			}
		}

		if err := h.installationRepo.SetRepositories(inst.ID, dbRepos); err != nil {
			log.Printf("Failed to save repos for installation %d: %v", inst.ID, err)
			continue
		}

		synced++
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"synced_count": synced,
		"total_count":  len(ghInstallations),
	})
}
