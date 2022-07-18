package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

func (r *queryResolver) Item(ctx context.Context, id string) (models.Item, error) {
	itemID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return models.MetadataModel{}, fmt.Errorf("invalid item id")
	}

	item, err := database.GetItemByID(uint(itemID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item")

		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// TODO; Properly handle all item types
	return helpers.GetItemFromItemMetadata(item), nil
}

func (r *queryResolver) Items(ctx context.Context, limit *int64, offset *int64, libraryID string) ([]models.Item, error) {
	parsedLibraryID, err := strconv.ParseUint(libraryID, 10, 64)
	if err != nil {
		return []models.Item{}, fmt.Errorf("invalid item id")
	}

	items, err := database.GetItemsFromLibrary(uint(parsedLibraryID), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	resultItems := make([]models.Item, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, helpers.GetItemFromItemMetadata(item))
	}

	return resultItems, nil
}

func (r *queryResolver) Children(ctx context.Context, limit *int64, offset *int64, item string) ([]models.Item, error) {
	parsedItemID, err := strconv.ParseUint(item, 10, 64)
	if err != nil {
		return []models.Item{}, fmt.Errorf("invalid item id")
	}

	items, err := database.GetChildrenFromItem(uint(parsedItemID), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	resultItems := make([]models.Item, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, helpers.GetItemFromItemMetadata(item))
	}

	return resultItems, nil
}

func (r *subscriptionResolver) OnItemAdded(ctx context.Context) (<-chan models.Item, error) {
	id := uuid.New().String()
	msg := make(chan models.Item, 1)

	go func() {
		<-ctx.Done()
		utils.SubsciptionsManager.Lock()
		delete(utils.SubsciptionsManager.ItemAddedObservers, id)
		utils.SubsciptionsManager.Unlock()
	}()
	utils.SubsciptionsManager.Lock()

	utils.SubsciptionsManager.ItemAddedObservers[id] = msg
	utils.SubsciptionsManager.Unlock()

	return msg, nil
}

func (r *subscriptionResolver) OnItemUpdated(ctx context.Context) (<-chan models.Item, error) {
	id := uuid.New().String()
	msg := make(chan models.Item, 1)

	go func() {
		<-ctx.Done()
		utils.SubsciptionsManager.Lock()
		delete(utils.SubsciptionsManager.ItemUpdatedObservers, id)
		utils.SubsciptionsManager.Unlock()
	}()
	utils.SubsciptionsManager.Lock()

	utils.SubsciptionsManager.ItemUpdatedObservers[id] = msg
	utils.SubsciptionsManager.Unlock()

	return msg, nil
}
