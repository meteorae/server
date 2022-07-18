package photos

import (
	"path/filepath"
	"strings"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/filter"
	"github.com/meteorae/meteorae-server/utils"
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

func Scan(path string, files, dirs *[]string, mediaList *[]models.Item, extensions []string, root string) {
	filter.Scan(path, files, dirs, mediaList, extensions, root)

	for _, filePath := range *files {
		file := filepath.Base(filePath)
		// Split filename and extension
		fileExtension := strings.ToLower(filepath.Ext(file))
		fileName := file[:len(file)-len(fileExtension)]

		if utils.IsStringInSlice(fileExtension, PhotoFileExtensions) {
			photo := models.Photo{
				Title: fileName,
				MetadataModel: &models.MetadataModel{
					Parts: []database.MediaPart{
						{
							FilePath: filepath.Join(root, path, filePath),
						},
					},
				},
				PhotoAlbum: models.PhotoAlbum{},
			}

			*mediaList = append(*mediaList, photo)
		} else {
			videoClip := models.VideoClip{
				Title: fileName,
				MetadataModel: &models.MetadataModel{
					Parts: []database.MediaPart{
						{
							FilePath: filepath.Join(root, path, filePath),
						},
					},
				},
				PhotoAlbum: models.PhotoAlbum{},
			}

			*mediaList = append(*mediaList, videoClip)
		}
	}
}
