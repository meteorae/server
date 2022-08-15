package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func (r *queryResolver) Latest(ctx context.Context, limit *int64) ([]*models.LatestResult, error) {
	latest := make([]*models.LatestResult, 0)

	libraries := database.GetLibraries()

	for _, library := range libraries {
		var latestItems []sdk.Item

		items, err := database.GetLatestItemsFromLibrary(*library, int(*limit))
		if err != nil {
			log.Err(err).Msgf("Failed to get latest items from library %d", library.ID)

			return nil, fmt.Errorf("failed to get latest items from library %d: %w", library.ID, err)
		}

		for i := range items {
			itemInfo, getInfoErr := metadata.GetInfoXML(items[i])
			if getInfoErr != nil {
				log.Err(getInfoErr).Msgf("Failed to get info XML for item %d", items[i].ID)

				return nil, fmt.Errorf("failed to get info XML for item %d: %w", items[i].ID, getInfoErr)
			}

			latestItems = append(latestItems, itemInfo)
		}

		latest = append(latest, &models.LatestResult{
			Library: library,
			Items:   latestItems,
		})
	}

	return latest, nil
}
