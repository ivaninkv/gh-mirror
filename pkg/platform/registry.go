package platform

import (
	"sync"

	"gh-mirror/pkg/models"
)

type Factory func() Platform

type Registry map[models.PlatformID]Factory

var (
	globalRegistry = make(Registry)
	registryMu     sync.RWMutex
)

func Register(id models.PlatformID, factory Factory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	globalRegistry[id] = factory
}

func Create(id models.PlatformID) (Platform, error) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	factory, exists := globalRegistry[id]
	if !exists {
		return nil, ErrPlatformNotFound
	}
	return factory(), nil
}

func RegisteredIDs() []models.PlatformID {
	registryMu.RLock()
	defer registryMu.RUnlock()
	ids := make([]models.PlatformID, 0, len(globalRegistry))
	for id := range globalRegistry {
		ids = append(ids, id)
	}
	return ids
}
