package movie

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/graph/model"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/stack"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

var (
	niceMatch = `(.+) [\(\[]([1-2][0-9]{4})[\)\]]`

	episodeRegex = []string{
		`(?P<show>.*?)[sS](?P<season>[0-9]+)[\._ ]*[eE](?P<ep>[0-9]+)[\._ ]*([- ]?[sS](?P<secondSeason>[0-9]+))?([- ]?[Ee+](?P<secondEp>[0-9]+))?`, // S03E04-E05
		`(?P<show>.*?)[sS](?P<season>[0-9]{2})[\._\- ]+(?P<ep>[0-9]+)`,                                                                             // S03-03
		`(?P<show>.*?)([^0-9]|^)(?P<season>(19[3-9][0-9]|20[0-5][0-9]|[0-9]{1,2}))[Xx](?P<ep>[0-9]+)((-[0-9]+)?[Xx](?P<secondEp>[0-9]+))?`,         // 3x03, 3x03-3x04, 3x03x04
		`(.*?)(^|[\._\- ])+(?P<season>sp)(?P<ep>[0-9]{2,3})([\._\- ]|$)+`,                                                                          // SP01 (Special 01, equivalent to S00E01)
	}
	standaloneTvRegex = `(.*?)( \(([0-9]+)\))? - ([0-9])+x([0-9]+)(-[0-9]+[Xx]([0-9]+))? - (.*)`
)

func GetName() string {
	return "Movie Scanner"
}

// TODO: Add ISO support.
func Scan(path string, files, dirs *[]string, mediaList *[]model.Item, extensions []string, root string) {
	video.Scan(path, files, dirs, mediaList, extensions, root)

	// Check for DVD rips.
	paths := filepath.SplitList(path)

	var videoTs string
	if utils.IsStringInSlice("video_ts.bup", *files) {
		videoTs = "video_ts.bup"
	}

	if len(paths) >= 1 && len(paths[0]) > 0 && videoTs != "" {
		log.Debug().Str("scanner", GetName()).Msgf("Found a DVD at %s", path)

		var (
			name string
			year int
		)

		// Figure out the name of the movie.
		if strings.ToLower(paths[len(paths)-1]) == "video_ts" && len(paths) >= 2 {
			// Easiest case, when the DVD rip is in a folder named after it's movie.
			name, year = video.CleanName(paths[len(paths)-2])
		} else {
			// Work up until we find something suitable.
			backwardsPaths := utils.ReverseSlice(paths)

			for _, p := range backwardsPaths {
				niceMatchRegex := regexp.MustCompile(niceMatch)
				if niceMatchRegex.MatchString(p) {
					name, year = video.CleanName(p)

					break
				}
			}

			if name == "" {
				// If we still don't have a name, just use the topmost path.
				name, year = video.CleanName(paths[0])
			}
		}

		movie := models.Movie{
			MetadataModel: &models.MetadataModel{
				Parts: []database.MediaPart{
					{
						FilePath: filepath.Join(root, path, videoTs),
					},
				},
			},
			Title:       name,
			ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
		}

		var (
			biggestFile string
			biggestSize int64
		)

		for _, file := range *files {
			fileSize, err := utils.GetFileSize(filepath.Join(root, path, file))
			if err != nil {
				log.Err(err).Str("scanner", GetName()).Msgf("Failed to get file size for %s", file)

				continue
			}

			if strings.ToLower(filepath.Ext(file)) == ".vob" && fileSize > biggestSize {
				biggestFile = file
				biggestSize = fileSize
			}
		}

		// Add the biggest part in order to get thumbnail/analysis/etc from it.
		if biggestFile != "" {
			movie.Parts = append(movie.Parts, database.MediaPart{
				FilePath: filepath.Join(root, path, biggestFile),
			})
		}

		if len(movie.Parts) > 0 {
			*mediaList = append(*mediaList, movie)
		}
	} else if len(paths) >= 3 && strings.ToLower(paths[len(paths)-1]) == "stream" &&
		strings.ToLower(paths[len(paths)-2]) == "bdmv" {
		log.Debug().Str("scanner", GetName()).Msgf("Found a Blu-ray at %s", path)

		name, year := video.CleanName(paths[len(paths)-3])

		movie := models.Movie{
			Title:       name,
			ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
		}

		for _, file := range *files {
			movie.Parts = append(movie.Parts, database.MediaPart{
				FilePath: filepath.Join(root, path, file),
			})
		}

		*mediaList = append(*mediaList, movie)
	} else {
		for _, file := range *files {
			name, year := video.CleanName(strings.TrimSuffix(file, filepath.Ext(file)))

			tv := false
			for _, episodePattern := range episodeRegex {
				episodeRegexp := regexp.MustCompile(episodePattern)
				if episodeRegexp.MatchString(file) {
					tv = true

					log.Debug().Str("scanner", GetName()).Msgf("Skipped %s due to matching a TV episode", file)

					break
				}
			}

			if !tv {
				movie := models.Movie{
					MetadataModel: &models.MetadataModel{
						Parts: []database.MediaPart{
							{
								FilePath: filepath.Join(root, path, file),
							},
						},
					},
					Title:       name,
					ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
				}

				*mediaList = append(*mediaList, movie)
			}
		}

		// If we have more than one media, attempt to match multi-part movies.
		if len(*mediaList) > 1 {
			stack.Scan(path, files, dirs, mediaList, extensions, root)
		}
	}

	var dirsToRemove []string

	for _, dir := range *dirs {
		standAloneMatchRegex := regexp.MustCompile(standaloneTvRegex)
		if standAloneMatchRegex.MatchString(dir) {
			dirsToRemove = append(dirsToRemove, dir)
		}
	}

	for _, dir := range dirsToRemove {
		*dirs = utils.RemoveStringFromSlice(dir, *dirs)
	}
}
