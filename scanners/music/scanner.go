package music

import (
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/audio"
	"github.com/rs/zerolog/log"
)

func GetName() string {
	return "Music Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]models.Item, extensions []string, root string) {
	audio.Scan(path, files, dirs, mediaList, extensions, root)

	audio.Process(path, files, dirs, mediaList, extensions, root)

	for _, media := range *mediaList {
		log.Debug().Msgf("Found media: %+v", media)
	}
}
