// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/sdk"
)

type Agent struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

// Authentication payload returned on successful login.
type AuthPayload struct {
	Token string         `json:"token"`
	User  *database.User `json:"user"`
}

type Image struct {
	Hash *string `json:"hash"`
	URL  *string `json:"url"`
}

type ItemsResult struct {
	Items []sdk.Item `json:"items"`
	Total *int64     `json:"total"`
}

type LatestResult struct {
	Library *database.Library `json:"library"`
	Items   []sdk.Item        `json:"items"`
}

type Scanner struct {
	Name       string `json:"name"`
	Identifier string `json:"identifier"`
}

type ServerInfo struct {
	Version    string `json:"version"`
	Onboarding bool   `json:"onboarding"`
}

type UpdateInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}