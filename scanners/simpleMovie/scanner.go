package simplemovie

import (
	"path/filepath"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/rs/zerolog/log"
)

func GetName() string {
	return "Simple Movie Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]database.ItemMetadata, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	video.Scan(path, files, dirs, mediaList, extensions, root)

	// Just add everything to the media list.
	for _, file := range *files {
		log.Debug().Str("scanner", GetName()).Msgf("Adding %s", file)
		name, year := video.CleanName(file)

		movie := database.ItemMetadata{
			Title:       name,
			ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			MediaParts: []database.MediaPart{
				{
					FilePath: filepath.Join(root, path, file),
				},
			},
			Type: database.MovieItem,
		}

		*mediaList = append(*mediaList, movie)
	}
}
