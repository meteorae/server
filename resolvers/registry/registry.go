package registry

import (
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/rs/zerolog/log"
)

// Defines the structure of a file resolver.
type Resolver interface {
	// Returns the name of the resolver.
	GetName() string
	// Returns whether the resolver supports the given library type.
	SupportsLibraryType(library database.Library) bool
	// Returns whether the resolver supports the given file type.
	SupportsFileType(filePath string, isDir bool) bool
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
func ResolveFile(mediaPart *database.MediaPart, library database.Library, isDir bool) error {
	log.Debug().Msgf("Resolving file %s", mediaPart.FilePath)

	for _, resolver := range Registry {
		log.Debug().Msgf("Checking resolver %s", resolver.GetName())

		if resolver.SupportsLibraryType(library) && resolver.SupportsFileType(mediaPart.FilePath, isDir) {
			log.Debug().Msgf("Resolving file %s with resolver %s", mediaPart.FilePath, resolver.GetName())
			err := resolver.Resolve(mediaPart, library)
			if err != nil {
				return fmt.Errorf("failed to resolve file: %w", err)
			}
		}
	}

	return nil
}
