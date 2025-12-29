package llm

import (
	"clothing/internal/entity"
	"fmt"
	"strings"
	"sync"
)

// ProviderConstructor is a function that creates an AIService from a provider config.
type ProviderConstructor func(provider *entity.DbProvider) (AIService, error)

// ProviderFactory manages AIService instances with registration and caching.
type ProviderFactory struct {
	registry map[string]ProviderConstructor
	cache    sync.Map
	media    MediaService
	mu       sync.RWMutex
}

var (
	globalFactory *ProviderFactory
	factoryOnce   sync.Once
)

// GetFactory returns the global ProviderFactory instance.
// It initializes the factory with default providers on first call.
func GetFactory() *ProviderFactory {
	factoryOnce.Do(func() {
		globalFactory = &ProviderFactory{
			registry: make(map[string]ProviderConstructor),
			media:    NewMediaService(),
		}
		// Register default providers with wrapper functions
		globalFactory.Register(entity.ProviderDriverOpenRouter, wrapOpenRouter)
		globalFactory.Register(entity.ProviderDriverGemini, wrapGeminiService)
		globalFactory.Register(entity.ProviderDriverAiHubMix, wrapAiHubMix)
		globalFactory.Register(entity.ProviderDriverDashscope, wrapDashscope)
		globalFactory.Register(entity.ProviderDriverFal, wrapFalAI)
		globalFactory.Register(entity.ProviderDriverVolcengine, wrapVolcengine)
	})
	return globalFactory
}

// Wrapper functions to convert concrete types to AIService interface
func wrapOpenRouter(provider *entity.DbProvider) (AIService, error) {
	return NewOpenRouter(provider)
}

func wrapGeminiService(provider *entity.DbProvider) (AIService, error) {
	return NewGeminiService(provider)
}

func wrapAiHubMix(provider *entity.DbProvider) (AIService, error) {
	return NewAiHubMix(provider)
}

func wrapDashscope(provider *entity.DbProvider) (AIService, error) {
	return NewDashscope(provider)
}

func wrapFalAI(provider *entity.DbProvider) (AIService, error) {
	return NewFalAI(provider)
}

func wrapVolcengine(provider *entity.DbProvider) (AIService, error) {
	return NewVolcengine(provider)
}

// Register adds a provider constructor to the registry.
func (f *ProviderFactory) Register(driver string, constructor ProviderConstructor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.registry[strings.ToLower(driver)] = constructor
}

// Get returns an AIService for the given provider.
// It caches instances by provider ID for reuse.
func (f *ProviderFactory) Get(provider *entity.DbProvider) (AIService, error) {
	if provider == nil {
		return nil, fmt.Errorf("provider config is nil")
	}

	cacheKey := provider.ID

	// Check cache first
	if cached, ok := f.cache.Load(cacheKey); ok {
		return cached.(AIService), nil
	}

	// Get driver
	driver := strings.ToLower(strings.TrimSpace(provider.Driver))
	if driver == "" {
		driver = strings.ToLower(strings.TrimSpace(provider.ID))
	}

	// Look up constructor
	f.mu.RLock()
	constructor, ok := f.registry[driver]
	f.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unsupported provider driver: %s", provider.Driver)
	}

	// Create new service
	service, err := constructor(provider)
	if err != nil {
		return nil, err
	}

	// Store in cache
	f.cache.Store(cacheKey, service)

	return service, nil
}

// Invalidate removes a cached provider instance.
// Call this when provider configuration changes.
func (f *ProviderFactory) Invalidate(providerID string) {
	f.cache.Delete(providerID)
}

// InvalidateAll clears all cached provider instances.
func (f *ProviderFactory) InvalidateAll() {
	f.cache.Range(func(key, _ any) bool {
		f.cache.Delete(key)
		return true
	})
}

// MediaService returns the shared MediaService instance.
func (f *ProviderFactory) MediaService() MediaService {
	return f.media
}

// ListDrivers returns all registered driver names.
func (f *ProviderFactory) ListDrivers() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	drivers := make([]string, 0, len(f.registry))
	for driver := range f.registry {
		drivers = append(drivers, driver)
	}
	return drivers
}

// NewService creates an AIService for a provider.
// This is the legacy function, kept for backward compatibility.
// Prefer using GetFactory().Get() for new code.
func NewService(provider *entity.DbProvider) (AIService, error) {
	return GetFactory().Get(provider)
}
