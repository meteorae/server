package api

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
	"github.com/rs/zerolog/log"
)

func (r *mutationResolver) Login(ctx context.Context, username string, password string) (*models.AuthPayload, error) {
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
		token, err := helpers.GenerateJwt(user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWT: %w", err)
		}

		return &models.AuthPayload{
			Token: token,
			User:  user,
		}, nil
	}

	return nil, errInvalidCredentials
}

func (r *mutationResolver) Register(ctx context.Context, username string, password string) (*models.AuthPayload, error) {
	user, err := database.CreateUser(username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := helpers.GenerateJwt(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT: %w", err)
	}

	return &models.AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (r *queryResolver) User(ctx context.Context, id string) (*database.User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}

	user, err := database.GetUserByID(uint(userID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")

		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *queryResolver) Users(ctx context.Context, limit *int64, offset *int64) ([]*database.User, error) {
	users := database.GetUsers()

	return users, nil
}

func (r *userResolver) ID(ctx context.Context, obj *database.User) (string, error) {
	return strconv.FormatUint(uint64(obj.ID), 10), nil //nolint:gomnd
}

// User returns models.UserResolver implementation.
func (r *Resolver) User() models.UserResolver { return &userResolver{r} }

type userResolver struct{ *Resolver }
