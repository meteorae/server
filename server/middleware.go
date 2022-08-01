package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	gqlHandler "github.com/99designs/gqlgen/graphql/handler"
	gqlExtension "github.com/99designs/gqlgen/graphql/handler/extension"
	gqlLru "github.com/99designs/gqlgen/graphql/handler/lru"
	gqlTransport "github.com/99designs/gqlgen/graphql/handler/transport"
	gqlPlayground "github.com/99designs/gqlgen/graphql/playground"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/handlers"
	"github.com/gorilla/websocket"
	"github.com/justinas/alice"
	"github.com/meteorae/meteorae-server/api"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/server/handlers/image/transcode"
	"github.com/meteorae/meteorae-server/server/handlers/library"
	"github.com/meteorae/meteorae-server/server/handlers/web"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

		validate, err := helpers.ValidateJwt(auth)
		if err != nil || !validate.Valid {
			log.Error().Msg("Request attempted with invalid token")
			http.Error(writer, "Invalid token", http.StatusForbidden)

			return
		}

		customClaim, _ := validate.Claims.(*helpers.JwtClaim)

		account, err := database.GetUserByID(customClaim.UserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				http.Error(writer, "Invalid token", http.StatusForbidden)

				return
			}
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		ctx := utils.GetContextWithUser(request.Context(), account)

		request = request.WithContext(ctx)
		next.ServeHTTP(writer, request)
	})
}

// TODO: Move this to GraphQL.
func setupHandler(writer http.ResponseWriter, request *http.Request) {
	userCount := database.GetUsersCount()

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
	router := gin.New()

	router.UseH2C = true

	recoverFunc := func(ctx context.Context, err interface{}) error {
		log.Error().Interface("error", err).Msg("A GraphQL error occurred")

		message := fmt.Sprintf("Internal system error. Error <%v>", err)

		return errors.New(message)
	}

	queryHandler := gqlHandler.New(models.NewExecutableSchema(models.Config{Resolvers: &api.Resolver{}}))
	queryHandler.SetRecoverFunc(recoverFunc)
	queryHandler.AddTransport(gqlTransport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		KeepAlivePingInterval: 10 * time.Second,
	})
	queryHandler.AddTransport(gqlTransport.Options{})
	queryHandler.AddTransport(gqlTransport.GET{})
	queryHandler.AddTransport(gqlTransport.POST{})
	queryHandler.AddTransport(gqlTransport.MultipartForm{
		MaxUploadSize: 1024 << 20, // 1GB
	})

	queryHandler.SetQueryCache(gqlLru.New(1000))
	queryHandler.Use(gqlExtension.Introspection{})

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

	err = web.EnsureWebClient()
	if err != nil {
		return nil, fmt.Errorf("failed to ensure web client: %w", err)
	}

	spa := web.SPAHandler{}

	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Debug().
			Str("method", httpMethod).
			Str("path", absolutePath).
			Str("handler", handlerName).
			Int("handlers", nuHandlers).
			Msg("")
	}

	router.Use(sentrygin.New(sentrygin.Options{
		// Send the panic to Gin's recovery handler
		Repanic: true,
	}))
	router.Use(cors.Default())

	router.Any("/graphql", gin.WrapH(queryHandler))
	router.GET("/playground", gin.WrapH(
		handlers.CompressHandler(
			gqlPlayground.Handler("GraphQL playground", "/query"))))
	router.GET("/image/transcode", gin.WrapH(
		handlers.CompressHandler(
			http.HandlerFunc(transcodeHandler.HTTPHandler))))
	router.GET("/library/{metadata}/{part}/file.{ext}", gin.WrapH(
		http.HandlerFunc(library.MediaPartHTTPHandler)))
	router.GET("/web", gin.WrapH(handlers.CompressHandler(spa)))

	// TODO: If we're not in HTTP/2, we should keep the connection alive

	h2s := http2.Server{}

	port := viper.GetInt("port")

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h2c.NewHandler(router, &h2s),
	}, nil
}
