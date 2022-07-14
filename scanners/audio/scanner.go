package audio

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/scanners/filter"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

var (
	AudioFileExtensions = []string{
		`mp3`,
		`m4a`,
		`m4b`,
		`flac`,
		`aac`,
		`rm`,
		`rma`,
		`mpa`,
		`wav`,
		`wma`,
		`ogg`,
		`mp2`,
		`mka`,
		`ac3`,
		`dts`,
		`ape`,
		`mpc`,
		`mp+`,
		`mpp`,
		`shn`,
		`oga`,
		`aiff`,
		`aif`,
		`wv`,
		`dsf`,
		`dsd`,
		`opus`,
	}
	variousArtists = []string{
		`va`,
		`v/a`,
		`various`,
		`various artists`,
		`various artist(s)`,
		`various artitsen`,
		`verschiedene`,
	}
)

func GetName() string {
	return "Audio Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]database.ItemMetadata, extensions []string, root string) {
	filter.Scan(path, files, dirs, mediaList, extensions, root)
}

func Process(path string, files, dirs *[]string, mediaList *[]database.ItemMetadata, extensions []string, root string) {
	if len(*files) == 0 {
		return
	}

	var albumTracks []database.ItemMetadata

	for _, file := range *files {
		artist, album, title, track, disc, albumArtist := getInfoFromTags(filepath.Join(root, path, file))

		if albumArtist != "" && utils.IsStringInSlice(strings.ToLower(albumArtist), variousArtists) {
			albumArtist = "Various Artists"
		}
		if artist == "" {
			artist = "[Unknown Artist]"
		}
		if album == "" {
			album = "[Unknown Album]"
		}

		title = strings.TrimSpace(title)

		track := database.ItemMetadata{
			Title: title,
		}
	}
}

func getInfoFromTags(file string) (string, string, string, int, int, string) {
	if strings.HasSuffix(strings.ToLower(file), "mp3") ||
		strings.HasSuffix(strings.ToLower(file), "mp4") ||
		strings.HasSuffix(strings.ToLower(file), "m4a") ||
		strings.HasSuffix(strings.ToLower(file), "m4b") ||
		strings.HasSuffix(strings.ToLower(file), "m4p") ||
		strings.HasSuffix(strings.ToLower(file), "ogg") ||
		strings.HasSuffix(strings.ToLower(file), "flac") {
		artist, album, title, track, disc, albumArtist := getTags(file)

		return artist, album, title, track, disc, albumArtist
	}

	// TODO: Figure out support for OggOpus and WMA

	return "", "", "", 0, 0, ""
}

func getTags(file string) (string, string, string, int, int, string) {
	mediaFile, err := os.Open(file)
	if err != nil {
		log.Err(err).Msgf("Failed to open file %s", file)

		return "", "", "", 0, 0, ""
	}
	defer mediaFile.Close()

	metadata, err := tag.ReadFrom(mediaFile)
	if err != nil {
		log.Err(err).Msgf("Failed to read tags from file %s", file)

		return "", "", "", 0, 0, ""
	}

	trackNumber, _ := metadata.Track()
	discNumber, _ := metadata.Disc()

	return metadata.Artist(), metadata.Album(), metadata.Title(), trackNumber, discNumber, metadata.AlbumArtist()
}
