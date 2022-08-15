package musicbrainz

import (
	"github.com/meteorae/meteorae-server/agents/musicbrainz/musicalbum"
	"github.com/meteorae/meteorae-server/sdk"
)

var AlbumPlugin sdk.AlbumAgent //nolint:gochecknoglobals // This is expected, since it's a plugin.

const (
	IDENTIFIER = "tv.meteorae.agents.musicbrainz"
	NAME       = "MusicBrainz"
)

type MusicBrainz struct {
	Identifier string
	Name       string
}

func (a MusicBrainz) GetIdentifier() string {
	return a.Identifier
}

func (a MusicBrainz) GetName() string {
	return a.Name
}

func (a MusicBrainz) GetAlbumAgent() *sdk.Agent {
	return &sdk.Agent{
		Identifier:           IDENTIFIER,
		Name:                 NAME,
		GetMetadataFunc:      musicalbum.GetMetadata,
		GetSearchResultsFunc: musicalbum.GetSearchResults,
	}
}

func init() {
	AlbumPlugin = MusicBrainz{
		Identifier: IDENTIFIER,
		Name:       NAME,
	}
}
