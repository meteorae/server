package graphql

import (
	"github.com/graphql-go/graphql"
)

var LibraryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Library",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"scannedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"libraryLocations": &graphql.Field{
			Type: graphql.NewList(LocationsType),
		},
	},
})

var LocationsType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Locations",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"rootPath": &graphql.Field{
			Type: graphql.String,
		},
		"available": &graphql.Field{
			Type: graphql.Boolean,
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"scannedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
	},
})
