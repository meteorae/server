package graphql

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/alexedwards/argon2id"
	"github.com/graphql-go/graphql"
	graphQlHandler "github.com/graphql-go/handler"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/filesystem/scanner"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	errAccessDenied          = errors.New("access denied")
	errFailedToParseArgument = errors.New("failed to parse argument")
	errInvalidCredentials    = errors.New("invalid credentials")

	pageLimit  = 10
	pageOffset = 10
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"metadata": &graphql.Field{
				Type: MovieType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					user := utils.GetUserFromContext(params.Context)
					if user == nil {
						return nil, errAccessDenied
					}

					metadataID, ok := params.Args["id"].(int)

					if ok {
						var metadata models.ItemMetadata
						database.DB.First(&metadata, metadataID)

						return metadata, nil
					}

					return nil, fmt.Errorf("failed to parse id: %w", errFailedToParseArgument)
				},
			},
			"allMetadata": &graphql.Field{
				Type: graphql.NewList(MovieType),
				Args: graphql.FieldConfigArgument{
					"limit": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: pageLimit,
					},
					"offset": &graphql.ArgumentConfig{
						Type:         graphql.Int,
						DefaultValue: pageOffset,
					},
					"libraryId": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					user := utils.GetUserFromContext(params.Context)
					if user == nil {
						return nil, errAccessDenied
					}

					limit, okLimit := params.Args["limit"].(int)
					if !okLimit {
						limit = 10
					}
					offset, okOffset := params.Args["offset"].(int)
					if !okOffset {
						offset = 0
					}

					libraryID, okLibraryID := params.Args["libraryId"].(int)
					if !okLibraryID {
						return nil, fmt.Errorf("failed to parse libraryId: %w", errFailedToParseArgument)
					}

					var metadata []models.ItemMetadata
					database.DB.Where("library_id = ?", libraryID).
						Limit(limit).
						Offset(offset).
						Preload("Cuts").
						Preload("Cuts.MediaPart").
						Find(&metadata)

					return metadata, nil
				},
			},
		},
	})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		// Scan a directory for media parts
		"addLibrary": &graphql.Field{
			Type:        LibraryType,
			Description: "Add a new library and start scanning for content",
			Args: graphql.FieldConfigArgument{
				"locations": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.String),
				},
				"type": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"language": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				user := utils.GetUserFromContext(params.Context)
				if user == nil {
					return nil, errAccessDenied
				}

				locations, locationOk := params.Args["locations"].([]interface{})
				if !locationOk {
					return nil, fmt.Errorf("failed to parse locations: %w", errFailedToParseArgument)
				}
				libraryTypeString, locationOk := params.Args["type"].(string)
				if !locationOk {
					return nil, fmt.Errorf("failed to parse type: %w", errFailedToParseArgument)
				}
				name, locationOk := params.Args["name"].(string)
				if !locationOk {
					return nil, fmt.Errorf("failed to parse name: %w", errFailedToParseArgument)
				}
				language, locationOk := params.Args["language"].(string)
				if !locationOk {
					return nil, fmt.Errorf("failed to parse language: %w", errFailedToParseArgument)
				}

				var libraryType models.LibraryType
				switch libraryTypeString {
				case string(models.MovieLibrary):
					libraryType = models.MovieLibrary
				case string(models.TVLibrary):
					libraryType = models.TVLibrary
				case string(models.MusicLibrary):
					libraryType = models.MusicLibrary
				case string(models.AnimeMovieLibrary):
					libraryType = models.AnimeMovieLibrary
				case string(models.AnimeTVLibrary):
					libraryType = models.AnimeTVLibrary
				}

				var libraryLocations []models.LibraryLocation
				for _, location := range locations {
					libraryLocations = append(libraryLocations, models.LibraryLocation{
						RootPath:  location.(string),
						Available: true,
					})
				}

				newLibrary := models.Library{
					Name:             name,
					Type:             libraryType,
					Language:         language,
					LibraryLocations: libraryLocations,
				}

				result := database.DB.Create(&newLibrary)

				if result.Error != nil {
					return nil, fmt.Errorf("failed to create library: %w", result.Error)
				}

				for _, location := range libraryLocations {
					err := ants.Submit(func() {
						scanner.ScanDirectory(location.RootPath, database.DB, newLibrary.Type)
					})
					if err != nil {
						log.Err(err).Msgf("Failed to schedule directory scan for %s", location.RootPath)
					}
				}

				return newLibrary, nil
			},
		},
		// Register a new user
		// TODO: This should be admin-only
		"registerAccount": &graphql.Field{
			Type:        LoginType,
			Description: "Register a new user",
			Args: graphql.FieldConfigArgument{
				"username": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				username, okUsername := params.Args["username"].(string)
				if !okUsername {
					return nil, fmt.Errorf("failed to parse username: %w", errFailedToParseArgument)
				}
				password, okPassword := params.Args["password"].(string)
				if !okPassword {
					return nil, fmt.Errorf("failed to parse password: %w", errFailedToParseArgument)
				}

				var account models.Account
				result := database.DB.Where("username = ?", username).First(&account)
				if result.Error == nil {
					// If the user exists, prevent registering one with the same name
					if errors.Is(result.Error, gorm.ErrRecordNotFound) {
						return nil, result.Error
					}
				}

				passwordHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
				if err != nil {
					log.Err(err).Msg("Could not create password hash")
				}

				newAccount := models.Account{
					Username: username,
					Password: passwordHash,
				}

				result = database.DB.Create(&newAccount)

				if result.Error != nil {
					return nil, fmt.Errorf("failed to create user: %w", result.Error)
				}

				token, err := helpers.GenerateJwt(params.Context, strconv.Itoa(int(newAccount.ID)))
				if err != nil {
					return nil, fmt.Errorf("failed to generate JWT: %w", err)
				}

				return map[string]interface{}{
					"token": token,
					"user":  newAccount,
				}, nil
			},
		},
		"login": &graphql.Field{
			Type:        LoginType,
			Description: "Login to an existing account",
			Args: graphql.FieldConfigArgument{
				"username": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"password": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				username, okUsername := params.Args["username"].(string)
				if !okUsername {
					return nil, fmt.Errorf("failed to parse username: %w", errFailedToParseArgument)
				}
				password, okPassword := params.Args["password"].(string)
				if !okPassword {
					return nil, fmt.Errorf("failed to parse password: %w", errFailedToParseArgument)
				}

				var account models.Account
				result := database.DB.Where("username = ?", username).First(&account)
				if result.Error != nil {
					if errors.Is(result.Error, gorm.ErrRecordNotFound) {
						return nil, errInvalidCredentials
					}

					return nil, result.Error
				}

				match, err := argon2id.ComparePasswordAndHash(password, account.Password)
				if err != nil {
					return nil, fmt.Errorf("failed to compare password: %w", err)
				}

				if !match {
					token, err := helpers.GenerateJwt(params.Context, strconv.Itoa(int(account.ID)))
					if err != nil {
						return nil, fmt.Errorf("failed to generate JWT: %w", err)
					}

					return map[string]interface{}{
						"token": token,
						"user":  account,
					}, nil
				}

				return nil, errInvalidCredentials
			},
		},
	},
})

var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    queryType,
	Mutation: mutationType,
})

func GetHandler() *graphQlHandler.Handler {
	return graphQlHandler.New(&graphQlHandler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})
}
