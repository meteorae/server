package movie

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/providers/movie"
	"github.com/meteorae/meteorae-server/resolvers/registry"
	"github.com/meteorae/meteorae-server/utils"
	PTN "github.com/middelink/go-parse-torrent-name"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

func init() {
	registry.Register(movieResolver)
}

var movieResolver registry.Resolver = Resolver{}

type Resolver struct{}

func (r Resolver) GetName() string {
	return "Movie"
}

func (r Resolver) SupportsLibraryType(library database.Library) bool {
	return library.Type == database.MovieLibrary
}

func (r Resolver) SupportsFileType(filePath string, isDir bool) bool {
	if isDir {
		return false
	}

	return isValidMovieFile(filePath)
}

func (r Resolver) Resolve(mediaPart *database.MediaPart, library database.Library) error {
	fileInfo, err := os.Stat(mediaPart.FilePath)
	if err != nil {
		return fmt.Errorf("could not stat %s: %w", mediaPart.FilePath, err)
	}

	if fileInfo.IsDir() {
		return nil
	}

	info, err := PTN.Parse(filepath.Base(mediaPart.FilePath))
	if err != nil {
		return fmt.Errorf("failed to parse movie name: %w", err)
	}

	item := database.ItemMetadata{
		Title:     info.Title,
		Type:      database.MovieItem,
		LibraryID: library.ID,
		Library:   library,
		MediaPart: *mediaPart,
	}

	err = database.CreateMovie(&item)
	if err != nil {
		return fmt.Errorf("could not resolve image metadata %s: %w", mediaPart.FilePath, err)
	}

	err = ants.Submit(func() {
		err := movie.GetInformation(&item, library)
		if err != nil {
			log.Err(err).Msgf("Failed to get movie information for %s: %s", mediaPart.FilePath, err)
		}
	})
	if err != nil {
		return fmt.Errorf("could not schedule image information job %s: %w", mediaPart.FilePath, err)
	}

	return nil
}

func isValidMovieFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.IsStringInSlice(ext, helpers.VideoFileExtensions)
}
