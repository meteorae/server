package utils

import (
	"context"

	"github.com/meteorae/meteorae-server/database/models"
)

type (
	authString string
)

func GetContextWithUser(ctx context.Context, user *models.Account) context.Context {
	return context.WithValue(ctx, authString("user"), *user)
}

func GetUserFromContext(ctx context.Context) *models.Account {
	raw, ok := ctx.Value(authString("user")).(models.Account)
	if !ok {
		return nil
	}

	return &raw
}
