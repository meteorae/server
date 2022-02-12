package registry

import (
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

// Defines the structure of a file resolver.
type Resolver interface {
	// Returns the name of the resolver.
	GetName() string
	// Returns whether the resolver supports the given library type.
	SupportsLibraryType(library database.Library) bool
	// Resolves the given media part.
	Resolve(mediaPart *database.MediaPart, library database.Library) error
}

var Registry []Resolver

// Registers a new file resolver.
func Register(resolver Resolver) {
	Registry = append(Registry, resolver)

	log.Debug().Msgf("Registered resolver %s", resolver.GetName())
}

// Schedules a file resolution job.
func ResolveFile(mediaPart *database.MediaPart, library database.Library) {
	err := ants.Submit(func() {
		log.Debug().Msgf("Resolving file %s", mediaPart.FilePath)
		err := resolveFileJob(mediaPart, library)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to resolve file %s", mediaPart.FilePath)
		}
	})
	if err != nil {
		log.Error().Err(err).Msg("Could not schedule file resolution job")
	}
}

func resolveFileJob(mediaPart *database.MediaPart, library database.Library) error {
	for _, resolver := range Registry {
		log.Debug().Msgf("Checking resolver %s", resolver.GetName())
		if resolver.SupportsLibraryType(library) {
			log.Debug().Msgf("Resolving file %s with resolver %s", mediaPart.FilePath, resolver.GetName())
			err := resolver.Resolve(mediaPart, library)
			if err != nil {
				return fmt.Errorf("failed to resolve file: %w", err)
			}
		}
	}

	return nil
}
