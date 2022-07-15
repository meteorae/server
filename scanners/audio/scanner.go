package audio

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/models"
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

func Scan(path string, files, dirs *[]string, mediaList *[]model.Item, extensions []string, root string) {
	filter.Scan(path, files, dirs, mediaList, extensions, root)
}

func Process(path string, files, dirs *[]string, mediaList *[]model.Item, extensions []string, root string) {
	if len(*files) == 0 {
		return
	}

	albumTracks := make([]models.Track, 0, len(*files))

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

		trackItem := models.Track{
			MetadataModel: &models.MetadataModel{
				Parts: []database.MediaPart{
					{
						FilePath: filepath.Join(root, path, file),
					},
				},
			},
			Title:       title,
			AlbumArtist: albumArtist,
			AlbumName:   album,
			Artist:      []string{artist},
			DiscIndex:   disc,
			Sequence:    track,
		}

		albumTracks = append(albumTracks, trackItem)
	}

	var (
		albumMap  = map[string][]models.Track{}
		artistMap = map[string]int{}
	)

	for _, track := range albumTracks {
		// Add all the album names to the album map
		if _, ok := albumMap[strings.ToLower(track.AlbumName)]; ok {
			albumMap[strings.ToLower(track.AlbumName)] = append(albumMap[strings.ToLower(track.AlbumName)], track)
		} else {
			albumMap[strings.ToLower(track.AlbumName)] = []models.Track{track}
		}

		// Count instances of identical artist names
		if _, ok := artistMap[strings.ToLower(track.Artist[0])]; ok {
			artistMap[strings.ToLower(track.Artist[0])]++
		} else {
			artistMap[strings.ToLower(track.Artist[0])] = 1
		}
	}

	// We want to see if this may be a Various Artist album.
	// To do so, we look at the artists, and figure out the most common one.
	// If the total number of tracks for the most popular artist is lower than a threshold,
	// we assume that this is a Various Artist album.
	maxArtistName := utils.RankMapStringInt(artistMap)[len(artistMap)-1]
	maxArtistCount := artistMap[maxArtistName]

	percentSameArtist := 0
	if len(albumTracks) > 0 {
		percentSameArtist = int(float64(maxArtistCount) / float64(len(albumTracks)) * 100) // nolint:gomnd
	}

	for album, tracks := range albumMap {
		sameAlbum := true
		sameArtist := true
		sameAlbumArtist := true
		previousAlbum := ""
		previousArtist := ""
		previousAlbumArtist := ""
		blankAlbumArtist := true

		for _, track := range tracks {
			if previousAlbum == "" {
				previousAlbum = track.AlbumName
			}

			if previousArtist == "" {
				previousArtist = track.Artist[0]
			}

			if previousAlbumArtist == "" {
				previousAlbumArtist = track.AlbumArtist
			}

			if strings.ToLower(previousAlbum) != strings.ToLower(track.AlbumName) {
				sameAlbum = false
			}

			if strings.ToLower(previousArtist) != strings.ToLower(track.Artist[0]) {
				sameArtist = false
			}

			if strings.ToLower(previousAlbumArtist) != strings.ToLower(track.AlbumArtist) {
				sameAlbumArtist = false
			}

			previousAlbum = track.AlbumName
			previousArtist = track.Artist[0]

			if track.AlbumArtist != "" && len(strings.TrimSpace(track.AlbumArtist)) > 0 {
				blankAlbumArtist = false
			}
		}

		var newArtist string

		if sameAlbum && !sameArtist && blankAlbumArtist {
			// Only replace if we have very high confidence
			if percentSameArtist < 90 {
				newArtist = "Various Artists"
			} else {
				newArtist = maxArtistName
			}
		}

		if sameArtist && sameAlbum && !sameAlbumArtist {
			for trackIndex, track := range tracks {
				track.AlbumArtist = newArtist

				albumMap[album][trackIndex] = track
			}
		}
	}

	for _, track := range albumTracks {
		*mediaList = append(*mediaList, track)
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
