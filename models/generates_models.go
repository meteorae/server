// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package models

import (
	"github.com/meteorae/meteorae-server/database"
)

// Authentication payload returned on successful login.
type AuthPayload struct {
	Token string         `json:"token"`
	User  *database.User `json:"user"`
}

type Image struct {
	Hash *string `json:"hash"`
	URL  *string `json:"url"`
}

type LatestResult struct {
	Library *database.Library        `json:"library"`
	Items   []*database.ItemMetadata `json:"items"`
}

type ServerInfo struct {
	Version    string `json:"version"`
	Onboarding bool   `json:"onboarding"`
}

type UpdateInfo struct {
	Version string `json:"version"`
	URL     string `json:"url"`
}
