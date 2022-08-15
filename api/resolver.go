package api

//go:generate go run github.com/99designs/gqlgen generate

import (
	"errors"
)

var (
	errInvalidCredentials = errors.New("invalid credentials")
	errInvalidItemID      = errors.New("invalid item id")
)

type Resolver struct{}
