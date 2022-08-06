package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers/metadata"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func (r *identifierResolver) Type(ctx context.Context, obj *sdk.Identifier) (int64, error) {
	return int64(obj.IdentifierType), nil
}

func (r *identifierResolver) Name(ctx context.Context, obj *sdk.Identifier) (string, error) {
	return obj.IdentifierType.String(), nil
}

func (r *identifierResolver) Value(ctx context.Context, obj *sdk.Identifier) (string, error) {
	return obj.Identifier, nil
}

func (r *movieResolver) ID(ctx context.Context, obj *sdk.Movie) (string, error) {
	return obj.ItemInfo.UUID.String(), nil
}

func (r *movieResolver) StartDate(ctx context.Context, obj *sdk.Movie) (*string, error) {
	if obj.ReleaseDate.IsZero() {
		return nil, nil
	}

	formattedDate := obj.ReleaseDate.Format("2006-01-02")

	return &formattedDate, nil
}

func (r *movieResolver) Thumb(ctx context.Context, obj *sdk.Movie) (*models.Image, error) {
	if len(obj.Thumb.Items) == 0 {
		return nil, nil
	}

	thumbURL := fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", obj.ID)

	_, hash := metadata.GetURIComponents(obj.Thumb.Items[0].Media)

	return &models.Image{
		Hash: &hash,
		URL:  &thumbURL,
	}, nil
}

func (r *movieResolver) Art(ctx context.Context, obj *sdk.Movie) (*models.Image, error) {
	if len(obj.Art.Items) == 0 {
		return nil, nil
	}

	artURL := fmt.Sprintf("/image/transcode?url=/metadata/%d/art", obj.ID)

	_, hash := metadata.GetURIComponents(obj.Art.Items[0].Media)

	return &models.Image{
		Hash: &hash,
		URL:  &artURL,
	}, nil
}

func (r *queryResolver) Item(ctx context.Context, id string) (sdk.Item, error) {
	itemID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid item id")
	}

	item, err := database.GetItemByID(uint(itemID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item")

		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	itemInfo, err := metadata.GetInfoXML(item)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item info")

		return nil, fmt.Errorf("failed to get item info: %w", err)
	}

	// TODO; Properly handle all item types
	return itemInfo, nil
}

func (r *queryResolver) Items(ctx context.Context, limit *int64, offset *int64, libraryID string) (*models.ItemsResult, error) {
	library := database.GetLibrary(libraryID)

	items, err := database.GetItemsFromLibrary(library, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	count, err := database.GetItemsCountFromLibrary(library)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items count")

		return nil, fmt.Errorf("failed to get items count: %w", err)
	}

	resultItems := make([]sdk.Item, 0, len(items))

	for _, item := range items {
		itemInfo, err := metadata.GetInfoXML(*item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get item info")

			return nil, fmt.Errorf("failed to get item info: %w", err)
		}

		resultItems = append(resultItems, itemInfo)
	}

	return &models.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *queryResolver) Children(ctx context.Context, limit *int64, offset *int64, item string) (*models.ItemsResult, error) {
	parsedItemID, err := strconv.ParseUint(item, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid item id")
	}

	items, err := database.GetChildrenFromItem(uint(parsedItemID), limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	parsedItemID, err = strconv.ParseUint(item, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid item id")
	}

	count, err := database.GetChildrenCountFromItem(uint(parsedItemID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items count")

		return nil, fmt.Errorf("failed to get items count: %w", err)
	}

	resultItems := make([]sdk.Item, 0, len(items))

	for _, item := range items {
		itemInfo, err := metadata.GetInfoXML(item)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get item info")

			return nil, fmt.Errorf("failed to get item info: %w", err)
		}

		resultItems = append(resultItems, itemInfo)
	}

	return &models.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *subscriptionResolver) OnItemAdded(ctx context.Context) (<-chan sdk.Item, error) {
	id := uuid.New().String()
	msg := make(chan sdk.Item, 1)

	go func() {
		<-ctx.Done()
		database.SubsciptionsManager.Lock()
		delete(database.SubsciptionsManager.ItemAddedObservers, id)
		database.SubsciptionsManager.Unlock()
	}()
	database.SubsciptionsManager.Lock()

	database.SubsciptionsManager.ItemAddedObservers[id] = msg
	database.SubsciptionsManager.Unlock()

	return msg, nil
}

func (r *subscriptionResolver) OnItemUpdated(ctx context.Context) (<-chan sdk.Item, error) {
	id := uuid.New().String()
	msg := make(chan sdk.Item, 1)

	go func() {
		<-ctx.Done()
		database.SubsciptionsManager.Lock()
		delete(database.SubsciptionsManager.ItemUpdatedObservers, id)
		database.SubsciptionsManager.Unlock()
	}()
	database.SubsciptionsManager.Lock()

	database.SubsciptionsManager.ItemUpdatedObservers[id] = msg
	database.SubsciptionsManager.Unlock()

	return msg, nil
}

// Identifier returns models.IdentifierResolver implementation.
func (r *Resolver) Identifier() models.IdentifierResolver { return &identifierResolver{r} }

// Movie returns models.MovieResolver implementation.
func (r *Resolver) Movie() models.MovieResolver { return &movieResolver{r} }

type (
	identifierResolver struct{ *Resolver }
	movieResolver      struct{ *Resolver }
)
