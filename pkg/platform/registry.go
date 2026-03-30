package platform

import "gh-mirror/pkg/models"

type Factory func() Platform

type Registry map[models.PlatformID]Factory

var globalRegistry = make(Registry)

func Register(id models.PlatformID, factory Factory) {
	globalRegistry[id] = factory
}

func Create(id models.PlatformID) (Platform, error) {
	factory, exists := globalRegistry[id]
	if !exists {
		return nil, ErrPlatformNotFound
	}
	return factory(), nil
}

func RegisteredIDs() []models.PlatformID {
	ids := make([]models.PlatformID, 0, len(globalRegistry))
	for id := range globalRegistry {
		ids = append(ids, id)
	}
	return ids
}
