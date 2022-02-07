package utils

import (
	"context"

	"github.com/meteorae/meteorae-server/database"
)

type (
	authString string
)

func GetContextWithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, authString("user"), *user)
}

func GetUserFromContext(ctx context.Context) *database.User {
	raw, ok := ctx.Value(authString("user")).(database.User)
	if !ok {
		return nil
	}

	return &raw
}
