package api

import (
	"errors"
)

var errInvalidCredentials = errors.New("invalid credentials")

type Resolver struct{}
