package platform

import "gh-mirror/pkg/models"

// Factory is a constructor function that returns a new Platform instance.
type Factory func() Platform

// Registry maps platform IDs to their factory functions.
type Registry map[models.PlatformID]Factory

var globalRegistry = make(Registry)

// Register adds a platform implementation to the global registry.
// Platform packages call this from their init() functions.
func Register(id models.PlatformID, factory Factory) {
	globalRegistry[id] = factory
}

// Create instantiates a new Platform by its ID.
// Returns ErrPlatformNotFound if the platform has not been registered.
func Create(id models.PlatformID) (Platform, error) {
	factory, exists := globalRegistry[id]
	if !exists {
		return nil, ErrPlatformNotFound
	}
	return factory(), nil
}

// RegisteredIDs returns the list of all registered platform IDs.
func RegisteredIDs() []models.PlatformID {
	ids := make([]models.PlatformID, 0, len(globalRegistry))
	for id := range globalRegistry {
		ids = append(ids, id)
	}
	return ids
}
