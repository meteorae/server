package audio

import (
	"path/filepath"

	"github.com/meteorae/meteorae-server/helpers"
	"github.com/meteorae/meteorae-server/utils"
)

func IsValidAudioFile(path string) bool {
	ext := filepath.Ext(path)

	return utils.IsStringInSlice(ext, helpers.AudioFileExtensions)
}
