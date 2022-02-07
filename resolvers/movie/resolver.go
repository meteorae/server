package movie

import (
	"fmt"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	tmdbProvider "github.com/meteorae/meteorae-server/providers/themoviedb"
	PTN "github.com/middelink/go-parse-torrent-name"
	"github.com/rs/zerolog/log"
)

func Resolve(mediaPart *database.MediaPart, library database.Library) error {
	log.Info().Msgf("Attempting to match %s", mediaPart.FilePath)

	if library.Type == database.MovieLibrary {
		movie, err := PTN.Parse(filepath.Base(mediaPart.FilePath))
		if err != nil {
			return fmt.Errorf("failed to parse movie name: %w", err)
		}

		log.Info().Msgf("Found movie: %s", movie.Title)

		// TODO: If we want to support multiple providers, we'll need to do this differently
		movieInfo, err := tmdbProvider.GetMovieInfoFromTmdb(movie, mediaPart, library)
		if err != nil {
			return fmt.Errorf("failed to get movie info: %w", err)
		}

		err = database.CreateMovie(movieInfo)
		if err != nil {
			return fmt.Errorf("failed to create movie: %w", err)
		}
	}

	return nil
}
