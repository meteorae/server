package graphql

import (
	"github.com/graphql-go/graphql"
)

var MovieType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Movie",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"title": &graphql.Field{
			Type: graphql.String,
		},
		"sortTitle": &graphql.Field{
			Type: graphql.String,
		},
		"originalTitle": &graphql.Field{
			Type: graphql.String,
		},
		"tagline": &graphql.Field{
			Type: graphql.String,
		},
		"releaseDate": &graphql.Field{
			Type: graphql.String,
		},
		"popularity": &graphql.Field{
			Type: graphql.Float,
		},
		"adult": &graphql.Field{
			Type: graphql.Boolean,
		},
		"originalLanguage": &graphql.Field{
			Type: graphql.String,
		},
		"budget": &graphql.Field{
			Type: graphql.Int,
		},
		"revenue": &graphql.Field{
			Type: graphql.Int,
		},
		"mediaPart": &graphql.Field{
			Type: MediaPartType,
		},
	},
})
