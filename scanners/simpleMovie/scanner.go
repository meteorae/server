package simplemovie

import (
	"path/filepath"
	"time"

	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func GetName() string {
	return "Video Files Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	video.Scan(path, files, dirs, mediaList, extensions, root)

	// Just add everything to the media list.
	for _, file := range *files {
		log.Debug().Str("scanner", GetName()).Msgf("Adding %s", file)
		name, year := video.CleanName(file)

		movie := sdk.Movie{
			ItemInfo: &sdk.ItemInfo{
				Parts: []string{
					filepath.Join(root, path, file),
				},
				Title:       name,
				ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		*mediaList = append(*mediaList, movie)
	}
}