package tv

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/meteorae/meteorae-server/scanners/stack"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
)

var (
	//nolint:lll
	episodeRegex = []string{
		`(?i)(?P<show>.*?)[sS](?P<season>[0-9]+)[\._ ]*[eE](?P<ep>[0-9]+)[\._ ]*([- ]?[sS](?P<secondSeason>[0-9]+))?([- ]?[Ee+](?P<secondEp>[0-9]+))?`, // S03E04-E05
		`(?i)(?P<show>.*?)[sS](?P<season>[0-9]{2})[\._\- ]+(?P<ep>[0-9]+)`,                                                                             // S03-03
		`(?i)(?P<show>.*?)([^0-9]|^)(?P<season>(19[3-9][0-9]|20[0-5][0-9]|[0-9]{1,2}))[Xx](?P<ep>[0-9]+)((-[0-9]+)?[Xx](?P<secondEp>[0-9]+))?`,         // 3x03, 3x03-3x04, 3x03x04
		`(?i)(.*?)(^|[\._\- ])+(?P<season>sp)(?P<ep>[0-9]{2,3})([\._\- ]|$)+`,                                                                          // SP01 (Special 01, equivalent to S00E01)
		// FIXME: Go doesn't do backreferences, so...?
		// `(?i)(?:.*?)([^0-9a-z])(?P<season>[0-9]{1,2})(?P<ep>[0-9]{2})([\.\-][0-9]+(?P<secondEp>[0-9]{2})([ \-_\.]|$)[\.\-]?)?(\1|$)`,                   // .602.
	}
	//nolint:lll
	dateRegex = []string{
		`(?i)(?P<year>[0-9]{4})[^0-9a-zA-Z()[\]]+(?P<month>[0-9]{2})[^0-9a-zA-Z()[\]]+(?P<day>[0-9]{2})([^0-9]|$)`,       // 2009-02-10
		`(?i)(?P<month>[0-9]{2})[^0-9a-zA-Z()[\]]+(?P<day>[0-9]{2})[^0-9a-zA-Z()[\]]+(?P<year>[0-9]{4})([^0-9a-zA-Z]|$)`, // 02-10-2009
	}

	seasonRegex = `.*?(?P<season>[0-9]+)$` // folder for a season

	onlyEpisodeRegex = []string{
		`(?i)(?P<ep>[0-9]{1,3})[\. -_]*of[\. -_]*[0-9]{1,3}`,      // 01 of 08
		`(?i)^(?P<ep>[0-9]{1,3})[^0-9]`,                           // 01 - Foo
		`(?i)e[a-z]*[ \.\-_]*(?P<ep>[0-9]{2,3})([^0-9c-uw-z%]|$)`, // Lorem Ipsum ep234
		`(?i).*?[ \.\-_](?P<ep>[0-9]{2,3})[^0-9c-uw-z%]+`,         // Lorem Ipsum - 04 - Dolor Sit Amet
		`(?i).*?[ \.\-_](?P<ep>[0-9]{2,3})$`,                      // Lorem Ipsum - 04
		`(?i).*?[^0-9x](?P<ep>[0-9]{2,3})$`,                       // Lorem707
		`(?i)^(?P<ep>[0-9]{1,3})$`,                                // 01
	}

	endsWithEpisode = []string{`(?i)[ ]*[0-9]{1,2}x[0-9]{1,3}$`, `(?i)[ ]*S[0-9]+E[0-9]+$`}
)

func GetName() string {
	return "Series Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	video.Scan(path, files, dirs, mediaList, extensions, root)

	// This uses strings.Split because filepath.SplitList erroneously splits on some characters, and not on others.
	// It would fail to split paths with a season folder, for example.
	paths := strings.Split(path, string(os.PathSeparator))
	shouldStack := true //nolint:ifshort // False positive

	if len(paths) == 1 && paths[0] == "." {
		// Top level directory
		for _, file := range *files {
			var (
				show       string
				season     string
				episode    int64
				endEpisode int64
			)

			baseFile := filepath.Base(file)

			for _, regex := range episodeRegex[0 : len(episodeRegex)-1] {
				compiledRegex := regexp.MustCompile(regex)

				matches := utils.FindNamedMatches(compiledRegex, baseFile)
				if len(matches) > 0 {
					if showMatch, ok := matches["show"]; ok {
						show = showMatch
					}

					if seasonMatch, ok := matches["season"]; ok {
						season = seasonMatch
					}

					if strings.ToLower(season) == "sp" {
						season = "0"
					}

					if episodeMatch, ok := matches["ep"]; ok {
						var err error

						episode, err = strconv.ParseInt(episodeMatch, 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting episode number to int")

							continue
						}

						endEpisode = episode
					}

					if secondEpMatch, ok := matches["secondEp"]; ok {
						var err error

						endEpisode, err = strconv.ParseInt(secondEpMatch, 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting episode number to int")

							continue
						}
					}

					name, year := video.CleanName(show)
					if name != "" {
						for i := episode; i <= endEpisode; i++ {
							seasonNumber, err := strconv.ParseInt(season, 10, 32)
							if err != nil {
								log.Err(err).Str("file", file).Msg("Error converting season number to int")

								// Force to season 1 if we fail to convert to an integer
								seasonNumber = 1
							}

							item := sdk.TVEpisode{
								ItemInfo: &sdk.ItemInfo{
									Parts:       []string{filepath.Join(root, path, file)},
									ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
								},
								SeriesTitle: name,
								Season:      int(seasonNumber),
								Episode:     int(i),
							}

							*mediaList = append(*mediaList, item)
						}
					}
				}
			}
		}
	} else if len(paths) > 0 && len(paths[0]) > 0 {
		done := false
		var (
			show         string
			season       string
			episode      int
			endEpisode   int
			year         int
			month        int
			day          int
			seasonNumber string
		)

		show, year = video.CleanName(paths[0])

		// Figure out which part of the path looks like a season
		if len(paths) >= 2 {
			season = paths[len(paths)-1]

			compiledSeasonRegex := regexp.MustCompile(seasonRegex)
			match := utils.FindNamedMatches(compiledSeasonRegex, season)

			seasonNumber = match["season"]
		}

		// Sometimes, episode names end up in the show name
		for _, regex := range endsWithEpisode {
			compiledEndsWithEpisodeRegex := regexp.MustCompile(regex)

			show = compiledEndsWithEpisodeRegex.ReplaceAllString(show, "")
		}

		for _, file := range *files {
			done = false

			// Split filename and extension
			baseFile := filepath.Base(file)
			extension := filepath.Ext(file)
			baseFile = strings.TrimSuffix(baseFile, extension)

			for _, reges := range dateRegex {
				compiledDateRegex := regexp.MustCompile(reges)

				matches := utils.FindNamedMatches(compiledDateRegex, baseFile)

				if len(matches) > 0 {
					parsedYear, err := strconv.ParseInt(matches["year"], 10, 32)
					if err != nil {
						log.Err(err).Str("file", file).Msg("Error converting year number to int")
					}

					year = int(parsedYear)

					parsedMonth, err := strconv.ParseInt(matches["month"], 10, 32)
					if err != nil {
						log.Err(err).Str("file", file).Msg("Error converting month number to int")
					}

					month = int(parsedMonth)

					parsedDay, err := strconv.ParseInt(matches["day"], 10, 32)
					if err != nil {
						log.Err(err).Str("file", file).Msg("Error converting day number to int")
					}

					day = int(parsedDay)

					item := sdk.TVEpisode{
						ItemInfo: &sdk.ItemInfo{
							Parts:       []string{filepath.Join(root, path, file)},
							ReleaseDate: time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC),
						},
						SeriesTitle: show,
						Season:      year,
					}

					*mediaList = append(*mediaList, item)

					done = true

					break
				}
			}

			if !done {
				_, cleanYear := video.CleanName(baseFile)
				if year == 0 && cleanYear != 0 {
					year = cleanYear
				}

				if cleanYear != 0 {
					file = strings.Replace(baseFile, strconv.Itoa(cleanYear), "", -1)
				}

				junkRegex := []string{
					`(?i)([hHx][\.]?264)[^0-9]`,
					`(?i)[^[0-9](720[pP])`,
					`(?i)[^[0-9](1080[pP])`,
					`(?i)[^[0-9](480[pP])`,
				}
				for _, regex := range junkRegex {
					compiledJunkRegex := regexp.MustCompile(regex)

					baseFile = compiledJunkRegex.ReplaceAllString(baseFile, "")
				}

				for _, regex := range episodeRegex {
					compiledEpisodeRegex := regexp.MustCompile(regex)

					matches := utils.FindNamedMatches(compiledEpisodeRegex, baseFile)
					if len(matches) > 0 {
						season = matches["season"]
						if strings.ToLower(season) == "sp" {
							season = "0"
						}

						parsedEpisode, err := strconv.ParseInt(matches["ep"], 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting episode number to int")
						}

						episode = int(parsedEpisode)

						if _, ok := matches["secondEp"]; ok {
							parsedEndEpisode, secondEpParseErr := strconv.ParseInt(matches["secondEp"], 10, 32)
							if secondEpParseErr != nil {
								log.Err(secondEpParseErr).Str("file", file).Msg("Error converting episode number to int")
							}

							endEpisode = int(parsedEndEpisode)
						}

						if regex == episodeRegex[len(episodeRegex)-1] {
							// Skip it if it looks like a movie
							movieSimilarityRegex := regexp.MustCompile(`.+ \([1-2][0-9]{3}\)`)
							if movieSimilarityRegex.MatchString(baseFile) {
								done = true

								break
							}

							if season == "0" {
								break
							}

							if seasonNumber != "" && seasonNumber != season {
								startsWithEpisodeRegex := regexp.MustCompile(`^[0-9]+[ -]`)

								if startsWithEpisodeRegex.MatchString(baseFile) {
									break
								}

								seasonNumber, seasonParseErr := strconv.ParseInt(season, 10, 32)
								if seasonParseErr != nil {
									log.Err(seasonParseErr).Str("file", file).Msg("Error converting season number to int")

									// Force to season 1 if we fail to convert to an integer
									seasonNumber = 1
								}

								// Treat the entire thing as an episode
								episode = int(episode + int(seasonNumber)*100)
								if endEpisode != 0 {
									endEpisode = int(endEpisode + int(seasonNumber)*100)
								}
							}
						}

						parsedSeasonNumber, err := strconv.ParseInt(season, 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting season number to int")

							// Force to season 1 if we fail to convert to an integer
							parsedSeasonNumber = 1
						}

						for i := episode; i <= endEpisode+1; i++ {
							item := sdk.TVEpisode{
								ItemInfo: &sdk.ItemInfo{
									Parts:       []string{filepath.Join(root, path, file)},
									ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
								},
								SeriesTitle: show,
								Season:      int(parsedSeasonNumber),
								Episode:     i,
							}

							*mediaList = append(*mediaList, item)
						}

						done = true

						break
					}
				}
			}

			if !done {
				fileName, fileYear := video.CleanName(baseFile)

				// If we didn't get a year before, use the one we got from the file.
				if year == 0 && fileYear != 0 {
					year = fileYear
				}

				for _, onlyEpisode := range onlyEpisodeRegex {
					onlyEpisodeCompiledRegex := regexp.MustCompile(onlyEpisode)

					matches := utils.FindNamedMatches(onlyEpisodeCompiledRegex, fileName)
					if len(matches) > 0 {
						parsedEpisode, err := strconv.ParseInt(matches["ep"], 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting episode number to int")
						}

						episode = int(parsedEpisode)

						if seasonNumber != "" {
							parsedSeasonNumber, seasonParseErr := strconv.ParseInt(seasonNumber, 10, 32)
							if seasonParseErr != nil {
								log.Err(seasonParseErr).Str("file", file).Msg("Error converting season number to int")

								// Force to season 1 if we fail to convert to an integer
								seasonNumber = "1"
							}

							if parsedEpisode >= 100 && parsedEpisode/100 == parsedSeasonNumber {
								episode = int(parsedEpisode % 100)
							}
						}

						// Standalone episodes can lead to false positive with "XX of YY". Avoid stacking them.
						if onlyEpisode == onlyEpisodeRegex[0] {
							shouldStack = false
						}

						parsedSeasonNumber, err := strconv.ParseInt(seasonNumber, 10, 32)
						if err != nil {
							log.Err(err).Str("file", file).Msg("Error converting season number to int")

							// Force to season 1 if we fail to convert to an integer
							seasonNumber = "1"
						}

						item := sdk.TVEpisode{
							ItemInfo: &sdk.ItemInfo{
								Parts:       []string{filepath.Join(root, path, file)},
								ReleaseDate: time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC),
							},
							SeriesTitle: show,
							Season:      int(parsedSeasonNumber),
							Episode:     episode,
						}

						*mediaList = append(*mediaList, item)

						done = true

						break
					}
				}
			}

			// If we still don't have anything at this point, just log it.
			if !done {
				log.Debug().Str("scanner", GetName()).Msgf("No results from %s", file)
			}
		}
	}

	if shouldStack {
		stack.Scan(path, files, dirs, mediaList, extensions, root)
	}
}
