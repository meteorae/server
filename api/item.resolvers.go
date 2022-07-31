package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/models"
	"github.com/rs/zerolog/log"
)

func (r *itemResolver) ID(ctx context.Context, obj *database.ItemMetadata) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

func (r *itemResolver) StartDate(ctx context.Context, obj *database.ItemMetadata) (*string, error) {
	if obj.ReleaseDate.IsZero() {
		return nil, nil
	}

	formattedDate := obj.ReleaseDate.Format("2006-01-02")

	return &formattedDate, nil
}

func (r *itemResolver) EndDate(ctx context.Context, obj *database.ItemMetadata) (*string, error) {
	if obj.EndDate.IsZero() {
		return nil, nil
	}

	formattedDate := obj.EndDate.Format("2006-01-02")

	return &formattedDate, nil
}

func (r *itemResolver) Artist(ctx context.Context, obj *database.ItemMetadata) (*database.ItemMetadata, error) {
	if !(obj.Type == database.MusicAlbumItem) {
		return nil, nil
	}

	// TODO: Use relationships to get artist

	artist, err := database.GetItemByID(obj.ParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get artist: %w", err)
	}

	return &artist, nil
}

func (r *itemResolver) Series(ctx context.Context, obj *database.ItemMetadata) (*database.ItemMetadata, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *itemResolver) Season(ctx context.Context, obj *database.ItemMetadata) (*database.ItemMetadata, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *itemResolver) Index(ctx context.Context, obj *database.ItemMetadata) (int64, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *itemResolver) Thumb(ctx context.Context, obj *database.ItemMetadata) (*models.Image, error) {
	if obj.Thumb == "" {
		return nil, nil
	}

	thumbURL := fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", obj.ID)

	return &models.Image{
		Hash: &obj.Thumb,
		URL:  &thumbURL,
	}, nil
}

func (r *itemResolver) Art(ctx context.Context, obj *database.ItemMetadata) (*models.Image, error) {
	if obj.Art == "" {
		return nil, nil
	}

	artURL := fmt.Sprintf("/image/transcode?url=/metadata/%d/art", obj.ID)

	return &models.Image{
		Hash: &obj.Art,
		URL:  &artURL,
	}, nil
}

func (r *itemResolver) Type(ctx context.Context, obj *database.ItemMetadata) (string, error) {
	return obj.Type.String(), nil
}

func (r *queryResolver) Item(ctx context.Context, id string) (*database.ItemMetadata, error) {
	itemID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return &database.ItemMetadata{}, fmt.Errorf("invalid item id")
	}

	item, err := database.GetItemByID(uint(itemID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item")

		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// TODO; Properly handle all item types
	return &item, nil
}

func (r *queryResolver) Items(ctx context.Context, limit *int64, offset *int64, libraryID string) (*models.ItemsResult, error) {
	library := database.GetLibrary(libraryID)

	items, err := database.GetItemsFromLibrary(library, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	count, err := database.GetItemsCountFromLibrary(library.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items count")

		return nil, fmt.Errorf("failed to get items count: %w", err)
	}

	resultItems := make([]*database.ItemMetadata, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, item)
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

	resultItems := make([]*database.ItemMetadata, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, &item)
	}

	return &models.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *subscriptionResolver) OnItemAdded(ctx context.Context) (<-chan *database.ItemMetadata, error) {
	id := uuid.New().String()
	msg := make(chan *database.ItemMetadata, 1)

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

func (r *subscriptionResolver) OnItemUpdated(ctx context.Context) (<-chan *database.ItemMetadata, error) {
	id := uuid.New().String()
	msg := make(chan *database.ItemMetadata, 1)

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

// Item returns models.ItemResolver implementation.
func (r *Resolver) Item() models.ItemResolver { return &itemResolver{r} }

type itemResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *itemResolver) IsRefreshing(ctx context.Context, obj *database.ItemMetadata) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *itemResolver) IsAnalyzing(ctx context.Context, obj *database.ItemMetadata) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}
