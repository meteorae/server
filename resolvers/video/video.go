package video

import (
	"path/filepath"

	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
)

func IsValidVideoFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.StringInSlice(ext, helpers.VideoFileExtensions)
}
