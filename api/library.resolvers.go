package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/filesystem/scanner"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/utils"
	ants "github.com/panjf2000/ants/v2"
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

func (r *mutationResolver) AddLibrary(ctx context.Context, typeArg string, name string, language string, locations []string) (*database.Library, error) {
	library, libraryLocations, err := database.CreateLibrary(name, language, typeArg, locations)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create library")

		return nil, fmt.Errorf("failed to create library: %w", err)
	}

	// TODO: Move this to a library manager
	for _, location := range libraryLocations {
		err := ants.Submit(func() {
			scanner.ScanDirectory(location.RootPath, *library)
		})
		if err != nil {
			log.Err(err).Msgf("Failed to schedule directory scan for %s", location.RootPath)
		}
	}

	for _, observer := range utils.SubsciptionsManager.LibraryAddedObservers {
		observer <- library
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

func (r *subscriptionResolver) OnLibraryAdded(ctx context.Context) (<-chan *database.Library, error) {
	id := uuid.New().String()
	msg := make(chan *database.Library, 1)

	go func() {
		<-ctx.Done()
		utils.SubsciptionsManager.Lock()
		delete(utils.SubsciptionsManager.LibraryAddedObservers, id)
		utils.SubsciptionsManager.Unlock()
	}()
	utils.SubsciptionsManager.Lock()

	utils.SubsciptionsManager.LibraryAddedObservers[id] = msg
	utils.SubsciptionsManager.Unlock()

	return msg, nil
}

// Library returns models.LibraryResolver implementation.
func (r *Resolver) Library() models.LibraryResolver { return &libraryResolver{r} }

type libraryResolver struct{ *Resolver }
