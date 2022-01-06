package movie

import (
	"fmt"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database/models"
	tmdbProvider "github.com/meteorae/meteorae-server/providers/themoviedb"
	PTN "github.com/middelink/go-parse-torrent-name"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Resolve(mediaPart *models.MediaPart, database *gorm.DB, libraryType models.LibraryType) error {
	log.Info().Msgf("Attempting to match %s", mediaPart.FilePath)

	if libraryType == models.MovieLibrary {
		movie, err := PTN.Parse(filepath.Base(mediaPart.FilePath))
		if err != nil {
			return fmt.Errorf("failed to parse movie name: %w", err)
		}

		log.Info().Msgf("Found movie: %s", movie.Title)

		// TODO: If we want to support multiple providers, we'll need to do this differently
		movieInfo, err := tmdbProvider.GetMovieInfoFromTmdb(movie, mediaPart)
		if err != nil {
			return fmt.Errorf("failed to get movie info: %w", err)
		}

		database.Create(movieInfo)
	}

	return nil
}
