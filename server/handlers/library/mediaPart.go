package library

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meteorae/meteorae-server/database"
	"github.com/rs/zerolog/log"
)

func MediaPartHTTPHandler(writer http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)

	metadataID := params["metadata"]
	mediaPartID := params["part"]

	mediaPart, err := database.GetMediaPartById(metadataID, mediaPartID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get media part")
		http.Error(writer, err.Error(), http.StatusInternalServerError)

		return
	}

	http.ServeFile(writer, request, mediaPart.FilePath)
}
