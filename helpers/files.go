package helpers

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/bmatcuk/doublestar/v4"
)

var baseDirectoryPermissions = os.FileMode(0o755)

var VideoFileExtensions = []string{
	".m4v",
	".3gp",
	".nsv",
	".ts",
	".ty",
	".strm",
	".rm",
	".rmvb",
	".ifo",
	".mov",
	".qt",
	".divx",
	".xvid",
	".bivx",
	".vob",
	".nrg",
	".img",
	".iso",
	".pva",
	".wmv",
	".asf",
	".asx",
	".ogm",
	".m2v",
	".avi",
	".bin",
	".dvr-ms",
	".mpg",
	".mpeg",
	".mp4",
	".mkv",
	".avc",
	".vp3",
	".svq3",
	".nuv",
	".viv",
	".dv",
	".fli",
	".flv",
	".001",
	".tp",
}

var AudioFileExtensions = []string{
	".nsv",
	".m4a",
	".flac",
	".aac",
	".strm",
	".pls",
	".rm",
	".mpa",
	".wav",
	".wma",
	".ogg",
	".opus",
	".mp3",
	".mp2",
	".mod",
	".amf",
	".669",
	".dmf",
	".dsm",
	".far",
	".gdm",
	".imf",
	".it",
	".m15",
	".med",
	".okt",
	".s3m",
	".stm",
	".sfx",
	".ult",
	".uni",
	".xm",
	".sid",
	".ac3",
	".dts",
	".cue",
	".aif",
	".aiff",
	".ape",
	".mac",
	".mpc",
	".mp+",
	".mpp",
	".shn",
	".wv",
	".nsf",
	".spc",
	".gym",
	".adplug",
	".adx",
	".dsp",
	".adp",
	".ymf",
	".ast",
	".afc",
	".hps",
	".xsp",
	".acc",
	".m4b",
	".oga",
	".dsf",
	".mka",
}

var BookFileExtensions = []string{
	".azw",
	".azw3",
	".cb7",
	".cbr",
	".cbt",
	".cbz",
	".epub",
	".mobi",
	".pdf",
}

var IgnoredFileGlobs = []string{
	"**/small.jpg",
	"**/albumart.jpg",

	// We have neither non-greedy matching or character group repetitions, working around that here.
	// https://github.com/dazinator/DotNet.Glob#patterns
	// .*/sample\..{1,5}
	"**/sample.?",
	"**/sample.??",
	"**/sample.???",  // Matches sample.mkv
	"**/sample.????", // Matches sample.webm
	"**/sample.?????",
	"**/*.sample.?",
	"**/*.sample.??",
	"**/*.sample.???",
	"**/*.sample.????",
	"**/*.sample.?????",
	"**/sample/*",

	// Directories
	"**/metadata/**",
	"**/metadata",
	"**/ps3_update/**",
	"**/ps3_update",
	"**/ps3_vprm/**",
	"**/ps3_vprm",
	"**/extrafanart/**",
	"**/extrafanart",
	"**/extrathumbs/**",
	"**/extrathumbs",
	"**/.actors/**",
	"**/.actors",
	"**/.wd_tv/**",
	"**/.wd_tv",
	"**/lost+found/**",
	"**/lost+found",

	// WMC temp recording directories that will constantly be written to
	"**/TempRec/**",
	"**/TempRec",
	"**/TempSBE/**",
	"**/TempSBE",

	// Synology
	"**/eaDir/**",
	"**/eaDir",
	"**/@eaDir/**",
	"**/@eaDir",
	"**/#recycle/**",
	"**/#recycle",

	// Qnap
	"**/@Recycle/**",
	"**/@Recycle",
	"**/.@__thumb/**",
	"**/.@__thumb",
	"**/$RECYCLE.BIN/**",
	"**/$RECYCLE.BIN",
	"**/System Volume Information/**",
	"**/System Volume Information",
	"**/.grab/**",
	"**/.grab",

	// Unix hidden files
	"**/.*",

	// Mac - if you ever remove the above.
	// "**/._*",
	// "**/.DS_Store",

	// thumbs.db
	"**/thumbs.db",

	// bts sync files
	"**/*.bts",
	"**/*.sync",
}

func ShouldIgnore(path string, d fs.DirEntry) bool {
	isMatched := false

	for _, ext := range IgnoredFileGlobs {
		match, err := doublestar.Match(ext, path)
		if err != nil {
			// If the glob fails, be safe and don't ignore the file
			break
		}

		if match {
			isMatched = true

			break
		}
	}

	return isMatched
}

func EnsurePathExists(path string) error {
	return fmt.Errorf("failed to ensure path exists: %w", os.MkdirAll(path, baseDirectoryPermissions))
}
