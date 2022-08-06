package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/agents"
	"github.com/meteorae/meteorae-server/database"
	filesystemScanner "github.com/meteorae/meteorae-server/filesystem/scanner"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners"
	"github.com/meteorae/meteorae-server/tasks"
	"github.com/rs/zerolog/log"
)

func (r *libraryResolver) ID(ctx context.Context, obj *database.Library) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *libraryResolver) Type(ctx context.Context, obj *database.Library) (string, error) {
	return obj.Type.String(), nil
}

func (r *libraryResolver) Locations(ctx context.Context, obj *database.Library) ([]string, error) {
	locations := make([]string, 0, len(obj.LibraryLocations))
	for _, location := range obj.LibraryLocations {
		locations = append(locations, location.RootPath)
	}

	return locations, nil
}

func (r *mutationResolver) AddLibrary(ctx context.Context, typeArg string, name string, language string, locations []string, scanner string, agent string) (*database.Library, error) {
	library, libraryLocations, err := database.CreateLibrary(name, language, typeArg, locations, scanner, agent)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create library")

		return nil, fmt.Errorf("failed to create library: %w", err)
	}

	err = tasks.LibraryScanQueue.Submit(func() {
		log.Info().Msgf("Scanning library %s", library.Name)

		for _, location := range libraryLocations {
			log.Info().Str("library", library.Name).Msgf("Scanning location %s", location.RootPath)

			filesystemScanner.ScanDirectory(location.RootPath, *library)
		}

		// After the scan is complete, queue up a metadata refresh.
		agents.RefreshLibraryMetadata(*library)
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to dispatch library scan task")
	}

	return library, nil
}

func (r *queryResolver) Library(ctx context.Context, id string) (*database.Library, error) {
	library := database.GetLibrary(id)

	return &library, nil
}

func (r *queryResolver) Libraries(ctx context.Context) ([]*database.Library, error) {
	libraries := database.GetLibraries()

	return libraries, nil
}

func (r *queryResolver) Scanners(ctx context.Context, libraryType string) ([]*models.Scanner, error) {
	scanners := scanners.GetScannerNamesForLibraryType(libraryType)

	return scanners, nil
}

func (r *queryResolver) Agents(ctx context.Context, libraryType string) ([]*models.Agent, error) {
	libraryTypeValue, err := database.LibraryTypeFromString(libraryType)
	if err != nil {
		return nil, fmt.Errorf("invalid library type: %w", err)
	}

	scanners := agents.GetAgentNamesForLibraryType(libraryTypeValue)

	return scanners, nil
}

func (r *subscriptionResolver) OnLibraryAdded(ctx context.Context) (<-chan *database.Library, error) {
	id := uuid.New().String()
	msg := make(chan *database.Library, 1)

	go func() {
		<-ctx.Done()
		database.SubsciptionsManager.Lock()
		delete(database.SubsciptionsManager.LibraryAddedObservers, id)
		database.SubsciptionsManager.Unlock()
	}()
	database.SubsciptionsManager.Lock()

	database.SubsciptionsManager.LibraryAddedObservers[id] = msg
	database.SubsciptionsManager.Unlock()

	return msg, nil
}

// Library returns models.LibraryResolver implementation.
func (r *Resolver) Library() models.LibraryResolver { return &libraryResolver{r} }

type libraryResolver struct{ *Resolver }
