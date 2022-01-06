package library

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/database/models"
	"github.com/rs/zerolog/log"
)

func MediaPartHTTPHandler(writer http.ResponseWriter, request *http.Request) {
	params := mux.Vars(request)

	metadataID := params["metadata"]
	mediaPartID := params["part"]

	var mediaPart models.MediaPart

	result := database.DB.Find(&mediaPart, "item_metadata_id = ? AND id = ?", metadataID, mediaPartID)
	if result.Error != nil {
		log.Error().Msg(result.Error.Error())
		http.Error(writer, result.Error.Error(), http.StatusInternalServerError)

		return
	}

	http.ServeFile(writer, request, mediaPart.FilePath)
}
