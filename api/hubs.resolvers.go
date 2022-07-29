package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/models"
	"github.com/rs/zerolog/log"
)

func (r *queryResolver) Latest(ctx context.Context, limit *int64) ([]*models.LatestResult, error) {
	var latest []*models.LatestResult

	libraries := database.GetLibraries()

	for _, library := range libraries {
		var latestItems []*database.ItemMetadata

		items, err := database.GetLatestItemsFromLibrary(*library, int(*limit))
		if err != nil {
			log.Err(err).Msgf("Failed to get latest items from library %d", library.ID)

			return nil, fmt.Errorf("failed to get latest items from library %d: %w", library.ID, err)
		}

		for i := range items {
			latestItems = append(latestItems, &items[i])
		}

		latest = append(latest, &models.LatestResult{
			Library: library,
			Items:   latestItems,
		})
	}

	return latest, nil
}
