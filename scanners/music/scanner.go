package music

import (
	"github.com/meteorae/meteorae-server/scanners/audio"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func GetName() string {
	return "Music Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	audio.Scan(path, files, dirs, mediaList, extensions, root)

	for _, media := range *mediaList {
		log.Debug().Msgf("Found media: %+v", media)
	}
}
