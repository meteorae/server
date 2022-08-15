package fanart

import (
	"github.com/meteorae/meteorae-server/agents/fanart/movie"
	"github.com/meteorae/meteorae-server/agents/fanart/tvshow"
	"github.com/meteorae/meteorae-server/sdk"
)

var (
	MoviePlugin sdk.MovieAgent //nolint:gochecknoglobals // This is expected, since it's a plugin.
	TVPlugin    sdk.TVAgent    //nolint:gochecknoglobals // This is expected, since it's a plugin.
)

const (
	IDENTIFIER = "tv.meteorae.agents.fanarttv"
	NAME       = "Fanart.tv"
)

type FanartTV struct {
	Identifier string
	Name       string
}

func (a FanartTV) GetIdentifier() string {
	return a.Identifier
}

func (a FanartTV) GetName() string {
	return a.Name
}

func (a FanartTV) GetMovieAgent() *sdk.Agent {
	return &sdk.Agent{
		Identifier:           IDENTIFIER,
		Name:                 NAME,
		GetMetadataFunc:      movie.GetMetadata,
		GetSearchResultsFunc: movie.GetSearchResults,
	}
}

func (a FanartTV) GetTVAgent() *sdk.Agent {
	return &sdk.Agent{
		Identifier:           IDENTIFIER,
		Name:                 NAME,
		GetMetadataFunc:      tvshow.GetMetadata,
		GetSearchResultsFunc: tvshow.GetSearchResults,
	}
}

func init() {
	MoviePlugin = FanartTV{
		Identifier: IDENTIFIER,
		Name:       NAME,
	}

	TVPlugin = FanartTV{
		Identifier: IDENTIFIER,
		Name:       NAME,
	}
}
