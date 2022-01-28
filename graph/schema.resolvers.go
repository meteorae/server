package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/filesystem/scanner"
	"github.com/meteorae/meteorae-server/graph/generated"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/helpers"
	ants "github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (r *libraryResolver) ID(ctx context.Context, obj *models.Library) (string, error) {
	return strconv.FormatUint(obj.ID, 10), nil //nolint:gomnd
}

func (r *libraryResolver) Type(ctx context.Context, obj *models.Library) (string, error) {
	return obj.Type.String(), nil
}

func (r *libraryResolver) Locations(ctx context.Context, obj *models.Library) ([]string, error) {
	locations := make([]string, 0, len(obj.LibraryLocations))
	for _, location := range obj.LibraryLocations {
		locations = append(locations, location.RootPath)
	}

	return locations, nil
}

func (r *mutationResolver) Login(ctx context.Context, username string,
	password string) (*model.AuthPayload, error) {
	var account models.User

	result := database.DB.Where("username = ?", username).First(&account)
	if result.Error != nil {
		log.Error().Err(result.Error).Msg("Failed to find user")

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errInvalidCredentials
		}

		return nil, result.Error
	}

	match, err := argon2id.ComparePasswordAndHash(password, account.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare password: %w", err)
	}

	if !match {
		token, err := helpers.GenerateJwt(strconv.Itoa(int(account.ID)))
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %w", err)
		}

		return &model.AuthPayload{
			Token: token,
			User:  &account,
		}, nil
	}

	return nil, errInvalidCredentials
}

func (r *mutationResolver) Register(ctx context.Context, username, password string) (*model.AuthPayload, error) {
	var account models.User

	result := database.DB.Where("username = ?", username).First(&account)
	if result.Error == nil {
		// If the user exists, prevent registering one with the same name
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		log.Err(err).Msg("Could not create password hash")
	}

	newAccount := models.User{
		Username: username,
		Password: passwordHash,
	}

	result = database.DB.Create(&newAccount)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	token, err := helpers.GenerateJwt(strconv.Itoa(int(newAccount.ID)))
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &model.AuthPayload{
		Token: token,
		User:  &newAccount,
	}, nil
}

func (r *mutationResolver) AddLibrary(ctx context.Context, typeArg, name,
	language string, locations []string) (*models.Library, error) {
	var libraryLocations []models.LibraryLocation //nolint:prealloc
	for _, location := range locations {
		libraryLocations = append(libraryLocations, models.LibraryLocation{
			RootPath:  location,
			Available: true,
		})
	}

	libraryType, err := models.LibraryTypeFromString(typeArg)
	if err != nil {
		return nil, fmt.Errorf("invalid library type: %w", err)
	}

	newLibrary := models.Library{
		Name:             name,
		Type:             libraryType,
		Language:         language,
		LibraryLocations: libraryLocations,
	}

	result := database.DB.Create(&newLibrary)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create library: %w", result.Error)
	}

	for _, location := range libraryLocations {
		err := ants.Submit(func() {
			scanner.ScanDirectory(location.RootPath, database.DB, newLibrary)
		})
		if err != nil {
			log.Err(err).Msgf("Failed to schedule directory scan for %s", location.RootPath)
		}
	}

	return &newLibrary, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*models.User, error) {
	var user models.User

	database.DB.First(&user, id)

	return &user, nil
}

func (r *queryResolver) Users(ctx context.Context, limit, offset *int64) (*model.UsersResult, error) {
	var users []*models.User

	database.DB.
		Limit(int(*limit)).
		Offset(int(*offset)).
		Find(&users)

	var count int64

	database.DB.Model(&models.User{}).Count(&count)

	return &model.UsersResult{
		Users: users,
		Total: &count,
	}, nil
}

func (r *queryResolver) Item(ctx context.Context, id string) (model.Metadata, error) {
	var item models.ItemMetadata

	database.DB.First(&item, id)

	// TODO; Properly handle all item types
	return &model.Movie{
		ID:          strconv.FormatUint(item.ID, 10), //nolint:gomnd
		Title:       item.Title,
		ReleaseDate: item.ReleaseDate.Unix(),
		Thumb:       fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", item.ID),
		Art:         fmt.Sprintf("/image/transcode?url=/metadata/%d/art", item.ID),
	}, nil
}

func (r *queryResolver) Items(ctx context.Context, limit, offset *int64, libraryID string) (*model.ItemsResult, error) {
	var items []*models.ItemMetadata

	database.DB.
		Limit(int(*limit)).
		Offset(int(*offset)).
		Where("library_id = ?", libraryID).
		Find(&items)

	var count int64

	database.DB.Model(&models.ItemMetadata{}).Where("library_id = ?", libraryID).Count(&count)

	resultItems := make([]model.Metadata, 0, len(items))
	for _, item := range items {
		resultItems = append(resultItems, &model.Movie{
			ID:          strconv.FormatUint(item.ID, 10), //nolint:gomnd
			Title:       item.Title,
			ReleaseDate: item.ReleaseDate.Unix(),
			Thumb:       fmt.Sprintf("/image/transcode?url=/metadata/%d/thumb", item.ID),
			Art:         fmt.Sprintf("/image/transcode?url=/metadata/%d/art", item.ID),
		})
	}

	return &model.ItemsResult{
		Items: resultItems,
		Total: &count,
	}, nil
}

func (r *queryResolver) Library(ctx context.Context, id string) (*models.Library, error) {
	var library models.Library

	database.DB.First(&library, id)

	return &library, nil
}

func (r *queryResolver) Libraries(ctx context.Context) (*model.LibrariesResult, error) {
	var libraries []*models.Library

	database.DB.Find(&libraries)

	var count int64

	database.DB.Model(&models.Library{}).Count(&count)

	return &model.LibrariesResult{
		Libraries: libraries,
		Total:     &count,
	}, nil
}

func (r *userResolver) ID(ctx context.Context, obj *models.User) (string, error) {
	return strconv.FormatUint(obj.ID, 10), nil //nolint:gomnd
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
