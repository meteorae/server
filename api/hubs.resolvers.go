package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

func (r *queryResolver) Latest(ctx context.Context, limit *int64) ([]*models.LatestResult, error) {
	var latest []*models.LatestResult

	libraries := database.GetLibraries()

	for _, library := range libraries {
		items, err := database.GetLatestItemsFromLibrary(*library, int(*limit))
		if err != nil {
			log.Err(err).Msgf("Failed to get latest items from library %d", library.ID)

			return nil, fmt.Errorf("failed to get latest items from library %d: %w", library.ID, err)
		}

		resultItems := make([]models.Item, 0, len(items))

		for _, item := range items {
			resultItems = append(resultItems, helpers.GetItemFromItemMetadata(item))
		}

		if len(items) > 0 {
			latest = append(latest, &models.LatestResult{
				Library: library,
				Items:   resultItems,
			})
		}
	}

	return latest, nil
}

func (r *subscriptionResolver) OnLatestItemAdded(ctx context.Context) (<-chan []*models.LatestResult, error) {
	id := uuid.New().String()
	msg := make(chan []*models.LatestResult, 1)

	go func() {
		<-ctx.Done()
		utils.SubsciptionsManager.Lock()
		delete(utils.SubsciptionsManager.LibraryAddedObservers, id)
		utils.SubsciptionsManager.Unlock()
	}()
	utils.SubsciptionsManager.Lock()

	utils.SubsciptionsManager.LatestItemsAddedObservers[id] = msg
	utils.SubsciptionsManager.Unlock()

	return msg, nil
}
