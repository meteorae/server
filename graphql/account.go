package graphql

import (
	"github.com/graphql-go/graphql"
)

var AccountType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Account",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"username": &graphql.Field{
			Type: graphql.String,
		},
		"createdAt": &graphql.Field{
			Type: graphql.DateTime,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.DateTime,
		},
	},
})

var LoginType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Login",
	Fields: graphql.Fields{
		"token": &graphql.Field{
			Type: graphql.String,
		},
		"user": &graphql.Field{
			Type: AccountType,
		},
	},
})
