package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/filesystem/scanner"
	"github.com/meteorae/meteorae-server/graph/generated"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/helpers"
	ants "github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

func (r *libraryResolver) ID(ctx context.Context, obj *database.Library) (string, error) {
	return strconv.FormatUint(obj.Id, 10), nil //nolint:gomnd
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

func (r *mutationResolver) Login(ctx context.Context, username, password string) (*model.AuthPayload, error) {
	user, err := database.GetUserByName(username)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	match, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare password: %w", err)
	}

	if match {
		token, err := helpers.GenerateJwt(user.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %w", err)
		}

		return &model.AuthPayload{
			Token: token,
			User:  user,
		}, nil
	}

	return nil, errInvalidCredentials
}

func (r *mutationResolver) Register(ctx context.Context, username, password string) (*model.AuthPayload, error) {
	user, err := database.CreateUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := helpers.GenerateJwt(user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &model.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (r *mutationResolver) AddLibrary(
	ctx context.Context,
	typeArg, name, language string,
	locations []string,
) (*database.Library, error) {
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

	return library, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*database.User, error) {
	user, err := database.GetUserById(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *queryResolver) Users(ctx context.Context, limit, offset *int64) (*model.UsersResult, error) {
	users, err := database.GetUsers()
	if err != nil {
		return nil, err
	}

	count, err := database.GetUsersCount()
	if err != nil {
		return nil, err
	}

	return &model.UsersResult{
		Users: users,
		Total: &count,
	}, nil
}

func (r *queryResolver) Item(ctx context.Context, id string) (model.Item, error) {
	item, err := database.GetItemById(id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get item")

		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	// TODO; Properly handle all item types
	return *helpers.GetItemFromItemMetadata(item), nil
}

func (r *queryResolver) Items(ctx context.Context, limit, offset *int64, libraryID string) (*model.ItemsResult, error) {
	items, err := database.GetItemsFromLibrary(libraryID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	count, err := database.GetItemsCountFromLibrary(libraryID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items count")

		return nil, fmt.Errorf("failed to get items count: %w", err)
	}

	resultItems := make([]model.Item, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, *helpers.GetItemFromItemMetadata(item))
	}

	return &model.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *queryResolver) Children(ctx context.Context, limit, offset *int64, item string) (*model.ItemsResult, error) {
	items, err := database.GetChildrenFromItem(item, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items")

		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	count, err := database.GetChildrenCountFromItem(item)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get items count")

		return nil, fmt.Errorf("failed to get items count: %w", err)
	}

	resultItems := make([]model.Item, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, *helpers.GetItemFromItemMetadata(item))
	}

	return &model.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *queryResolver) Library(ctx context.Context, id string) (*database.Library, error) {
	library, err := database.GetLibraryById(id)
	if err != nil {
		return nil, err
	}

	return &library, nil
}

func (r *queryResolver) Libraries(ctx context.Context) (*model.LibrariesResult, error) {
	libraries, err := database.GetLibraries()
	if err != nil {
		return nil, err
	}

	count, err := database.GetLibrariesCount()
	if err != nil {
		return nil, err
	}

	return &model.LibrariesResult{
		Libraries: libraries,
		Total:     &count,
	}, nil
}

func (r *queryResolver) Latest(ctx context.Context, limit *int64) ([]*model.LatestResult, error) {
	var latest []*model.LatestResult

	libraries, err := database.GetLibraries()
	if err != nil {
		return nil, err
	}

	for _, library := range libraries {
		items, err := database.GetLatestItemsFromLibrary(library.Id, int(*limit))
		if err != nil {
			log.Err(err).Msgf("Failed to get latest items from library %d", library.Id)

			return nil, fmt.Errorf("failed to get latest items from library %d: %w", library.Id, err)
		}

		resultItems := make([]model.Item, 0, len(items))

		for _, item := range items {
			resultItems = append(resultItems, *helpers.GetItemFromItemMetadata(item))
		}

		if len(items) > 0 {
			latest = append(latest, &model.LatestResult{
				Library: library,
				Items:   resultItems,
			})
		}
	}

	return latest, nil
}

func (r *userResolver) ID(ctx context.Context, obj *database.User) (string, error) {
	return obj.Id, nil
}

// Library returns generated.LibraryResolver implementation.
func (r *Resolver) Library() generated.LibraryResolver { return &libraryResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type (
	libraryResolver  struct{ *Resolver }
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
	userResolver     struct{ *Resolver }
)
