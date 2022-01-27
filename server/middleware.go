package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/meteorae/meteorae-server/graphql"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/server/handlers/image/transcode"
	"github.com/meteorae/meteorae-server/server/handlers/library"
	"github.com/meteorae/meteorae-server/server/handlers/web"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var (
	writeTimeout = 15 * time.Second
	readTimeout  = 15 * time.Second
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hlog.FromRequest(request).Info()

		next.ServeHTTP(writer, request)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		auth := request.Header.Get("Authorization")

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
		_, err := writer.Write([]byte("true"))
		if err != nil {
			log.Error().Msg(err.Error())
		}
	} else {
		_, err := writer.Write([]byte("false"))
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}
}

func GetWebServer() (*http.Server, error) {
	// Setup webserver and serve GraphQL handler
	router := mux.NewRouter()

	graphQlHandler := graphql.GetHandler()

	transcodeHandler, err := transcode.NewImageHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to create image handler: %w", err)
	}

	loggingHandler := alice.New()
	loggingHandler = loggingHandler.Append(hlog.NewHandler(log.Logger))
	loggingHandler = loggingHandler.Append(hlog.AccessHandler(
		func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("")
		}))
	loggingHandler = loggingHandler.Append(hlog.RemoteAddrHandler("ip"))
	loggingHandler = loggingHandler.Append(hlog.UserAgentHandler("user_agent"))
	loggingHandler = loggingHandler.Append(hlog.RefererHandler("referer"))
	loggingHandler = loggingHandler.Append(hlog.RequestIDHandler("req_id", "Request-Id"))

	spa := web.SPAHandler{}

	router.Handle("/setup", loggingHandler.Then(http.HandlerFunc(setupHandler))).Methods("GET")
	router.Handle("/graphql", loggingHandler.Then(graphQlHandler))
	router.Handle("/image/transcode", loggingHandler.Then(http.HandlerFunc(transcodeHandler.HTTPHandler)))
	router.Handle("/library/{metadata}/{part}/file.{ext}",
		loggingHandler.Then(http.HandlerFunc(library.MediaPartHTTPHandler)))
	router.PathPrefix("/").Handler(loggingHandler.Then(spa))
	router.Use(LoggingMiddleware)
	router.Use(AuthMiddleware)

	return &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:42000",
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}, nil
}
