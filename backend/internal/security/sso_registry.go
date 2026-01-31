package security

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// SSORegistry manages SSO providers for organizations
type SSORegistry struct {
	mu       sync.RWMutex
	repo     SSOConfigRepository
	cache    map[string][]*SSOProviderConfig // organizationID -> providers
	cacheTTL int64                           // cache TTL in seconds (0 = no caching)
}

// SSOConfigRepository defines the interface for SSO configuration storage
type SSOConfigRepository interface {
	// Provider CRUD
	Create(ctx context.Context, config *SSOProviderConfig) error
	GetByID(ctx context.Context, id string) (*SSOProviderConfig, error)
	Update(ctx context.Context, config *SSOProviderConfig) error
	Delete(ctx context.Context, id string) error

	// List and query
	ListByOrganization(ctx context.Context, orgID string) ([]*SSOProviderConfig, error)
	GetByOrganizationAndType(ctx context.Context, orgID string, providerType SSOProviderType) (*SSOProviderConfig, error)
	GetActiveProviders(ctx context.Context, orgID string) ([]*SSOProviderConfig, error)

	// Attribute mappings
	SaveAttributeMappings(ctx context.Context, providerID string, mappings []AttributeMapping) error
	GetAttributeMappings(ctx context.Context, providerID string) ([]AttributeMapping, error)
	DeleteAttributeMappings(ctx context.Context, providerID string) error
}

// NewSSORegistry creates a new SSO registry
func NewSSORegistry(repo SSOConfigRepository) *SSORegistry {
	return &SSORegistry{
		repo:  repo,
		cache: make(map[string][]*SSOProviderConfig),
	}
}

// RegisterProvider adds a new SSO provider for an organization
func (r *SSORegistry) RegisterProvider(ctx context.Context, config *SSOProviderConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Create in repository
	if err := r.repo.Create(ctx, config); err != nil {
		return err
	}

	// Save attribute mappings if present
	if len(config.AttributeMappings) > 0 {
		if err := r.repo.SaveAttributeMappings(ctx, config.ID, config.AttributeMappings); err != nil {
			return err
		}
	}

	// Invalidate cache
	delete(r.cache, config.OrganizationID)

	return nil
}

// GetProvider retrieves an SSO provider by ID
func (r *SSORegistry) GetProvider(ctx context.Context, providerID string) (*SSOProviderConfig, error) {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, errors.New("provider not found")
	}

	// Load attribute mappings
	mappings, err := r.repo.GetAttributeMappings(ctx, providerID)
	if err != nil {
		return nil, err
	}
	provider.AttributeMappings = mappings

	return provider, nil
}

// UpdateProvider updates an SSO provider
func (r *SSORegistry) UpdateProvider(ctx context.Context, config *SSOProviderConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Update in repository
	if err := r.repo.Update(ctx, config); err != nil {
		return err
	}

	// Update attribute mappings
	if err := r.repo.DeleteAttributeMappings(ctx, config.ID); err != nil {
		return err
	}
	if len(config.AttributeMappings) > 0 {
		if err := r.repo.SaveAttributeMappings(ctx, config.ID, config.AttributeMappings); err != nil {
			return err
		}
	}

	// Invalidate cache
	delete(r.cache, config.OrganizationID)

	return nil
}

// DeleteProvider removes an SSO provider
func (r *SSORegistry) DeleteProvider(ctx context.Context, providerID string, orgID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Delete attribute mappings first
	if err := r.repo.DeleteAttributeMappings(ctx, providerID); err != nil {
		return err
	}

	// Delete provider
	if err := r.repo.Delete(ctx, providerID); err != nil {
		return err
	}

	// Invalidate cache
	delete(r.cache, orgID)

	return nil
}

// ListProviders returns all SSO providers for an organization, sorted by priority
func (r *SSORegistry) ListProviders(ctx context.Context, orgID string) ([]*SSOProviderConfig, error) {
	r.mu.RLock()
	if cached, ok := r.cache[orgID]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	providers, err := r.repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Sort by priority
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	// Cache result
	r.mu.Lock()
	r.cache[orgID] = providers
	r.mu.Unlock()

	return providers, nil
}

// GetActiveProviders returns only enabled/active SSO providers for an organization
func (r *SSORegistry) GetActiveProviders(ctx context.Context, orgID string) ([]*SSOProviderConfig, error) {
	providers, err := r.repo.GetActiveProviders(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Sort by priority
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	return providers, nil
}

// GetProviderByType returns a provider of a specific type for an organization
func (r *SSORegistry) GetProviderByType(ctx context.Context, orgID string, providerType SSOProviderType) (*SSOProviderConfig, error) {
	return r.repo.GetByOrganizationAndType(ctx, orgID, providerType)
}

// EnableProvider enables an SSO provider
func (r *SSORegistry) EnableProvider(ctx context.Context, providerID string) error {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	if provider == nil {
		return errors.New("provider not found")
	}

	provider.Enabled = true
	provider.Status = SSOProviderStatusActive

	return r.UpdateProvider(ctx, provider)
}

// DisableProvider disables an SSO provider
func (r *SSORegistry) DisableProvider(ctx context.Context, providerID string) error {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	if provider == nil {
		return errors.New("provider not found")
	}

	provider.Enabled = false
	provider.Status = SSOProviderStatusInactive

	return r.UpdateProvider(ctx, provider)
}

// SetProviderPriority updates the display priority of a provider
func (r *SSORegistry) SetProviderPriority(ctx context.Context, providerID string, priority int) error {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	if provider == nil {
		return errors.New("provider not found")
	}

	provider.Priority = priority

	return r.UpdateProvider(ctx, provider)
}

// SetProviderError marks a provider as having an error
func (r *SSORegistry) SetProviderError(ctx context.Context, providerID string, errMsg string) error {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	if provider == nil {
		return errors.New("provider not found")
	}

	provider.Status = SSOProviderStatusError
	provider.LastError = errMsg

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.repo.Update(ctx, provider); err != nil {
		return err
	}

	delete(r.cache, provider.OrganizationID)
	return nil
}

// ClearProviderError clears a provider's error status
func (r *SSORegistry) ClearProviderError(ctx context.Context, providerID string) error {
	provider, err := r.repo.GetByID(ctx, providerID)
	if err != nil {
		return err
	}
	if provider == nil {
		return errors.New("provider not found")
	}

	provider.LastError = ""
	if provider.Enabled {
		provider.Status = SSOProviderStatusActive
	} else {
		provider.Status = SSOProviderStatusInactive
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.repo.Update(ctx, provider); err != nil {
		return err
	}

	delete(r.cache, provider.OrganizationID)
	return nil
}

// InvalidateCache clears the cache for an organization
func (r *SSORegistry) InvalidateCache(orgID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.cache, orgID)
}

// InvalidateAllCache clears all cached data
func (r *SSORegistry) InvalidateAllCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string][]*SSOProviderConfig)
}

// GetProviderSummary returns a summary of SSO providers for an organization
func (r *SSORegistry) GetProviderSummary(ctx context.Context, orgID string) (*SSOProviderSummary, error) {
	providers, err := r.ListProviders(ctx, orgID)
	if err != nil {
		return nil, err
	}

	summary := &SSOProviderSummary{
		OrganizationID: orgID,
		TotalProviders: len(providers),
		ProviderTypes:  make(map[SSOProviderType]int),
	}

	for _, p := range providers {
		summary.ProviderTypes[p.Type]++
		if p.Enabled && p.Status == SSOProviderStatusActive {
			summary.ActiveProviders++
		}
		if p.Status == SSOProviderStatusError {
			summary.ErrorProviders++
		}
	}

	return summary, nil
}

// SSOProviderSummary provides a summary of SSO providers for an organization
type SSOProviderSummary struct {
	OrganizationID  string                   `json:"organization_id"`
	TotalProviders  int                      `json:"total_providers"`
	ActiveProviders int                      `json:"active_providers"`
	ErrorProviders  int                      `json:"error_providers"`
	ProviderTypes   map[SSOProviderType]int  `json:"provider_types"`
}

// ReorderProviders reorders providers by setting their priorities
func (r *SSORegistry) ReorderProviders(ctx context.Context, orgID string, providerIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, id := range providerIDs {
		provider, err := r.repo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if provider == nil {
			continue
		}
		if provider.OrganizationID != orgID {
			return errors.New("provider does not belong to organization")
		}

		provider.Priority = i + 1
		if err := r.repo.Update(ctx, provider); err != nil {
			return err
		}
	}

	delete(r.cache, orgID)
	return nil
}
