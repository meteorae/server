package video

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/meteorae/meteorae-server/scanners/filter"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/meteorae/meteorae-server/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	minSampleSize = 300 * 1024 * 1024 // 300MB
)

var (
	VideoFileExtensions = []string{
		"3g2",
		"3gp",
		"asf",
		"asx",
		"avc",
		"avi",
		"avs",
		"bivx",
		"bup",
		"divx",
		"dv",
		"dvr-ms",
		"evo",
		"fli",
		"flv",
		"m2ts",
		"m2v",
		"m4v",
		"mkv",
		"mov",
		"mp4",
		"mpeg",
		"mpg",
		"mts",
		"nsv",
		"nuv",
		"ogm",
		"ogv",
		"pva",
		"qt",
		"rm",
		"rmvb",
		"stp",
		"svq3",
		"strm",
		"ts",
		"ty",
		"vdr",
		"viv",
		"vob",
		"vp3",
		"wmv",
		"wtv",
		"xsp",
		"xvid",
		"webm",
	}

	ignoreSamplesRegex = []string{
		`(?i)[-\._]sample`,
		`(?i)sample[-\._]`,
	}
	ignoreTrailersRegex = []string{`-trailer\.`}
	ignoreExtrasRegex   = []string{
		`(?i)^trailer.?$`,
		`(?i)-deleted\.`,
		`(?i)-behindthescenes\.`,
		`(?i)-interview\.`,
		`(?i)-scene\.`,
		`(?i)-featurette\.`,
		`(?i)-short\.`,
		`(?i)-other\.`,
	}
	ignoreExtrasStarsWithRegex = []string{`(?i)^movie-trailer.*`}
	ignoreDirectoriesRegex     = []string{
		`(?i)\\bextras?\\b`,
		`(?i)!?samples?`,
		`(?i)bonus`,
		`(?i).*bonus disc.*`,
		`(?i)bdmv`,
		`(?i)video_ts`,
		`(?i)^interview.?$`,
		`(?i)^scene.?$`,
		`(?i)^trailer.?$`,
		`(?i)^deleted.?(scene.?)?$`,
		`(?i)^behind.?the.?scenes$`,
		`(?i)^featurette.?$`,
		`(?i)^short.?$`,
		`(?i)^other.?$`,
	}
	ignoreSuffixes = []string{`.dvdmedia`}

	sourceMap = map[string][]string{
		"bluray":    {"bdrc", "bdrip", "bluray", "bd", "brrip", "hdrip", "hddvd", "hddvdrip"},
		"cam":       {"cam"},
		"dvd":       {"ddc", "dvdrip", "dvd", "r1", "r3", "r5"},
		"retail":    {"retail"},
		"dtv":       {"dsr", "dsrip", "hdtv", "pdtv", "ppv"},
		"stv":       {"stv", "tvrip"},
		"screener":  {"bdscr", "dvdscr", "dvdscreener", "scr", "screener"},
		"svcd":      {"svcd"},
		"vcd":       {"vcd"},
		"telecine":  {"tc", "telecine"},
		"telesync":  {"ts", "telesync"},
		"web":       {"webrip", "web-dl"},
		"workprint": {"wp", "workprint"},
	}

	sources      = []string{}
	sourceSlices = maps.Values(sourceMap)

	audioRegex = []string{
		`(?i)([^0-9])5\.1[ ]*ch(.)`,
		`(?i)([^0-9])5\.1([^0-9]?)`,
		`(?i)([^0-9])7\.1[ ]*ch(.)`,
		`(?i)([^0-9])7\.1([^0-9])`,
	}
	edition      = []string{"se"}
	formatTokens = []string{
		"ac3",
		"divx",
		"fragment",
		"limited",
		"ogg",
		"ogm",
		"ntsc",
		"pal",
		"ps3avchd",
		"r1",
		"r3",
		"r5",
		"720i",
		"720p",
		"1080i",
		"1080p",
		"remux",
		"x264",
		"xvid",
		"vorbis",
		"aac",
		"dts",
		"fs",
		"ws",
		"1920x1080",
		"1280x720",
		"h264",
		"h",
		"264",
		"prores",
		"uhd",
		"2160p",
		"truehd",
		"atmos",
		"hevc",
	}
	miscTokens = []string{
		"cd1",
		"cd2",
		"1cd",
		"2cd",
		"custom",
		"internal",
		"repack",
		"read.nfo",
		"readnfo",
		"nfofix",
		"proper",
		"rerip",
		"dubbed",
		"subbed",
		"extended",
		"unrated",
		"xxx",
		"nfo",
		"dvxa",
		"web",
	}
	subsTokens = []string{"multi", "multisubs"}

	yearRegex = regexp.MustCompile(`(?i)\b(((?:19[0-9]|20[0-9])[0-9]))\b`)

	bracketsRegex  = regexp.MustCompile(`(?i)\[[^\]]+\]`)
	tokenRegex     = regexp.MustCompile(`(?i)([^ \-_\.\(\)+]+)`)
	separatorRegex = regexp.MustCompile(`(?i)[\.\-_\(\)+]+`)
)

func init() {
	for _, sourceSlice := range sourceSlices {
		sources = append(sources, sourceSlice...)
	}
}

func GetName() string {
	return "Video Scanner"
}

// CleanName takes a string representing a video name and returns a cleaned string along with a year.
func CleanName(name string) (string, int) {
	// Always ensure the name is lowercase
	name = strings.ToLower(name)

	var year int64 = 0

	// Some movies have a year in the title, so we match all years and take the rightmost one.
	yearMatch := yearRegex.FindAllString(name, -1)

	if len(yearMatch) > 0 && yearMatch[len(yearMatch)-1] != "" {
		var yearParseErr error
		year, yearParseErr = strconv.ParseInt(yearMatch[len(yearMatch)-1], 10, 32)
		if yearParseErr != nil {
			log.Debug().Str("date", yearMatch[len(yearMatch)-1]).Err(yearParseErr).Msg("Failed to parse year")
		}

		// Remove the year from the name, to allow us to break early later on.
		// Anything after the year should be garbage.
		name = strings.Replace(name, yearMatch[len(yearMatch)-1], "*yearBreak*", -1)
	}

	// Remove everything in brackets
	name = bracketsRegex.ReplaceAllString(name, "")

	// Remove useless suffixes
	for _, suffix := range ignoreSuffixes {
		name = strings.TrimSuffix(name, suffix)
	}

	// Remove the audio specs
	for _, audioRegex := range audioRegex {
		name = regexp.MustCompile(audioRegex).ReplaceAllString(name, " ")
	}

	// Convert the remainder into tokens
	tokens := tokenRegex.FindAllString(name, -1)

	cleanTokens := []string{}

	// Remove all separators from the token list
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if !separatorRegex.MatchString(token) {
			cleanTokens = append(cleanTokens, strings.ToLower(token))
		}
	}

	// Remove the garbage tokens and create a new clean name
	garbage := append(subsTokens, miscTokens...)
	garbage = append(garbage, edition...)
	garbage = append(garbage, formatTokens...)
	garbage = append(garbage, sources...)
	garbage = append(garbage, VideoFileExtensions...)

	// Keep track of the tokens we've seen before
	seenTokens := map[string]bool{}
	tokenBitmap := []bool{}

	// The title is usually at the start, so start at the end since that's where
	// the garbage tokens probably are.
	for i := len(cleanTokens) - 1; i >= 0; i-- {
		token := cleanTokens[i]

		if utils.IsStringInSlice(token, garbage) && !utils.IsStringInSlice(token, maps.Keys(seenTokens)) {
			tokenBitmap = append([]bool{false}, tokenBitmap...)
			seenTokens[token] = true
		} else {
			tokenBitmap = append([]bool{true}, tokenBitmap...)
		}
	}

	// Strip out all the garbage tokens.
	// Heuristics is simple: if we encounter 2+ BADs after encountering a GOOD,
	// take out the rest (even if they aren't BAD).
	countGood := 0
	countBad := 0

	finalTokens := []string{}

	for i, tokenValue := range tokenBitmap {
		good := tokenValue

		// If there is only a few tokens, we do not remove anything.
		// This avoids removing the title of a movie.
		if len(tokenBitmap) <= 2 {
			good = true
		}

		if good && countBad < 2 {
			if strings.ToLower(cleanTokens[i]) == "*yearbreak*" {
				// If the year is not the first token, we can skip the rest.
				if i == 0 {
					continue
				} else {
					break
				}
			} else {
				finalTokens = append(finalTokens, cleanTokens[i])
			}
		} else if !good && cleanTokens[i] == "web" {
			// Web-dl gets split up during tokenization, so we need to check for it manually.
			if i+1 < len(cleanTokens) && cleanTokens[i+1] == "dl" {
				i += 1 //nolint:ineffassign
			}
		}

		if good {
			countGood += 1
		} else {
			countBad += 1
		}
	}

	// If we took everything out, use the first token. We don't want an empty name.
	if len(finalTokens) == 0 && len(cleanTokens) > 0 {
		finalTokens = append(finalTokens, cleanTokens[0])
	}

	cleanName := strings.Join(finalTokens, " ")

	// Some movies have multiple languages or versions in the name using "aka", get only the first one
	reg := regexp.MustCompile("(.*) aka .*")
	cleanAkaName := reg.FindStringSubmatch(cleanName)

	if len(cleanAkaName) > 0 {
		cleanName = cleanAkaName[1]
	}

	// TODO: Support multiple languages
	titleCaser := cases.Title(language.English)

	return titleCaser.String(cleanName), int(year)
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	filter.Scan(path, files, dirs, mediaList, extensions, root)

	filesToRemove := []string{}

	for _, file := range *files {
		// Remove samples only if they're smaller than 300MB
		for _, sample := range ignoreSamplesRegex {
			sampleRegex := regexp.MustCompile(sample)
			if sampleRegex.MatchString(file) {
				fileSize, err := utils.GetFileSize(file)
				if err != nil {
					log.Err(err).Msg("Could not get file size")
				}

				if fileSize < minSampleSize {
					filesToRemove = append(filesToRemove, file)
				}
			}
		}

		// Remove trailer files
		for _, trailer := range ignoreTrailersRegex {
			trailerRegex := regexp.MustCompile(trailer)
			if trailerRegex.MatchString(file) {
				filesToRemove = append(filesToRemove, file)
			}
		}

		// Remove extras - These are handled by another scanner
		for _, ignoreExtra := range ignoreExtrasRegex {
			extraRegex := regexp.MustCompile(ignoreExtra)
			if extraRegex.MatchString(file) {
				filesToRemove = append(filesToRemove, file)
			}
		}

		// Remove files that start with common extras patterns
		for _, extrasStartsWith := range ignoreExtrasStarsWithRegex {
			startsWithRegex := regexp.MustCompile(extrasStartsWith)
			if startsWithRegex.MatchString(file) {
				filesToRemove = append(filesToRemove, file)
			}
		}

		// Remove duplicates and process the files to remove
		filesToRemove = utils.RemoveDuplicatesFromSlice(filesToRemove)
		for _, fileToRemove := range filesToRemove {
			*files = utils.RemoveStringFromSlice(fileToRemove, *files)
		}

		// At the top level, we don't remove any directories in this scanner.
		// But when deeper, we filter for directories we don't want.
		ignoreDirs := []string{}
		if path != "" {
			ignoreDirs = ignoreDirectoriesRegex
		}

		dirsToRemove := []string{}

		for _, dir := range *dirs {
			baseDir := filepath.Base(dir)

			for _, ignoreDir := range ignoreDirs {
				ignoreDirRegex := regexp.MustCompile(ignoreDir)
				if ignoreDirRegex.MatchString(baseDir) {
					dirsToRemove = append(dirsToRemove, dir)
				}
			}
		}

		for _, dirToRemove := range dirsToRemove {
			*dirs = utils.RemoveStringFromSlice(dirToRemove, *dirs)
		}
	}
}

func GetSource(name string) string {
	wordsRegex := regexp.MustCompile(`[^a-zA-Z0-9']+`)
	words := wordsRegex.Split(name, -1)

	for _, word := range words {
		word = strings.ToLower(word)
		if utils.IsStringInSlice(word, sources) {
			return word
		}
	}

	return ""
}
