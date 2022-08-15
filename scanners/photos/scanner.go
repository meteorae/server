package photos

import (
	"path/filepath"
	"strings"

	"github.com/meteorae/meteorae-server/scanners/filter"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

var PhotoFileExtensions = []string{
	`png`,
	`jpg`,
	`jpeg`,
	`bmp`,
	`gif`,
	`ico`,
	`tif`,
	`tiff`,
	`tga`,
	`pcx`,
	`dng`,
	`nef`,
	`cr2`,
	`crw`,
	`orf`,
	`arw`,
	`erf`,
	`3fr`,
	`dcr`,
	`x3f`,
	`mef`,
	`raf`,
	`mrw`,
	`pef`,
	`sr2`,
	`mpo`,
	`jps`,
	`rw2`,
	`jp2`,
	`j2k`,
}

func GetName() string {
	return "Photo Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	filter.Scan(path, files, dirs, mediaList, extensions, root)

	for _, filePath := range *files {
		file := filepath.Base(filePath)
		// Split filename and extension
		fileExtension := strings.ToLower(filepath.Ext(file))
		fileName := file[:len(file)-len(fileExtension)]

		if utils.IsStringInSlice(fileExtension, PhotoFileExtensions) {
			photo := sdk.Image{
				ItemInfo: &sdk.ItemInfo{
					Title: fileName,
					Parts: []string{
						filepath.Join(root, path, filePath),
					},
				},
			}

			*mediaList = append(*mediaList, photo)
		} else {
			videoClip := sdk.Video{
				ItemInfo: &sdk.ItemInfo{
					Title: fileName,
					Parts: []string{filepath.Join(root, path, filePath)},
				},
			}

			*mediaList = append(*mediaList, videoClip)
		}
	}
}
