package themoviedb

import (
	"github.com/meteorae/meteorae-server/agents/themoviedb/movie"
	"github.com/meteorae/meteorae-server/agents/themoviedb/tvshow"
	"github.com/meteorae/meteorae-server/sdk"
)

var (
	MoviePlugin sdk.MovieAgent
	TVPlugin    sdk.TVAgent
)

const (
	IDENTIFIER = "tv.meteorae.agents.tmdb"
	NAME       = "The Movie Database"
)

type TheMovieDB struct {
	Identifier string
	Name       string
}

func (a TheMovieDB) GetIdentifier() string {
	return a.Identifier
}

func (a TheMovieDB) GetName() string {
	return a.Name
}

func (a TheMovieDB) GetMovieAgent() *sdk.Agent {
	return &sdk.Agent{
		Identifier:           IDENTIFIER,
		Name:                 NAME,
		GetMetadataFunc:      movie.GetMetadata,
		GetSearchResultsFunc: movie.GetSearchResults,
	}
}

func (a TheMovieDB) GetTVAgent() *sdk.Agent {
	return &sdk.Agent{
		Identifier:           IDENTIFIER,
		Name:                 NAME,
		GetMetadataFunc:      tvshow.GetMetadata,
		GetSearchResultsFunc: tvshow.GetSearchResults,
	}
}

func init() {
	MoviePlugin = TheMovieDB{
		Identifier: IDENTIFIER,
		Name:       NAME,
	}

	TVPlugin = TheMovieDB{
		Identifier: IDENTIFIER,
		Name:       NAME,
	}
}
