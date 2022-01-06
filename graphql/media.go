package graphql

import (
	"github.com/graphql-go/graphql"
)

var MediaPartType = graphql.NewObject(graphql.ObjectConfig{
	Name: "MediaPart",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"filePath": &graphql.Field{
			Type: graphql.String,
		},
		"hash": &graphql.Field{
			Type: graphql.String,
		},
		"openSubtitleHash": &graphql.Field{
			Type: graphql.Boolean,
		},
		"aniDBCRC": &graphql.Field{
			Type: graphql.String,
		},
		"acoustID": &graphql.Field{
			Type: graphql.String,
		},
		"size": &graphql.Field{
			Type: graphql.Int,
		},
		"duration": &graphql.Field{
			Type: graphql.Int,
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"deletedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
	},
})
