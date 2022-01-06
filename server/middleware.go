package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/graphql"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/server/handlers/image/transcode"
	"github.com/meteorae/meteorae-server/server/handlers/library"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	writeTimeout = 15 * time.Second
	readTimeout  = 15 * time.Second
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Debug().Msg(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		auth := request.Header.Get("Authorization")

		// If there's no authorization header, we don't need to do anything
		if auth == "" {
			log.Debug().Msg("No authorization header")
			next.ServeHTTP(writer, request)

			return
		}

		bearer := "Bearer "
		auth = auth[len(bearer):]

		validate, err := helpers.ValidateJwt(context.Background(), auth)
		if err != nil || !validate.Valid {
			log.Error().Msg("Request attempted with invalid token")
			http.Error(writer, "Invalid token", http.StatusForbidden)

			return
		}

		customClaim, _ := validate.Claims.(*helpers.JwtClaim)

		var account models.Account
		result := database.DB.Where("id = ?", customClaim.UserID).First(&account)
		log.Debug().Msgf("Request from %s", account.Username)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				http.Error(writer, "Invalid token", http.StatusForbidden)

				return
			}
			http.Error(writer, result.Error.Error(), http.StatusInternalServerError)
		}

		ctx := utils.GetContextWithUser(request.Context(), &account)

		request = request.WithContext(ctx)
		next.ServeHTTP(writer, request)
	})
}

func setupHandler(writer http.ResponseWriter, request *http.Request) {
	var userCount int64

	result := database.DB.Model(&models.Account{}).Count(&userCount)
	if result.Error != nil {
		http.Error(writer, result.Error.Error(), http.StatusInternalServerError)
	}

	if userCount == 0 {
		writer.Write([]byte("true"))
	} else {
		writer.Write([]byte("false"))
	}
}

func GetWebServer() (*http.Server, error) {
	// Setup webserver and serve GraphQL handler
	router := mux.NewRouter()

	graphQlHandler := graphql.GetHandler()

	transcodeHandler, err := transcode.NewImageHandler()
	if err != nil {
		return nil, err
	}

	router.Handle("/setup", http.HandlerFunc(setupHandler)).Methods("GET")
	router.Handle("/graphql", graphQlHandler)
	router.Handle("/image/transcode", http.HandlerFunc(transcodeHandler.HTTPHandler))
	router.Handle("/library/{metadata}/{part}/file.{ext}", http.HandlerFunc(library.MediaPartHTTPHandler))
	router.Use(LoggingMiddleware)
	router.Use(AuthMiddleware)

	return &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:42000",
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}, nil
}
