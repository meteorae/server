package scanners

import (
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/audio"
	"github.com/meteorae/meteorae-server/scanners/filter"
	"github.com/meteorae/meteorae-server/scanners/movie"
	"github.com/meteorae/meteorae-server/scanners/music"
	"github.com/meteorae/meteorae-server/scanners/photos"
	simpleMovie "github.com/meteorae/meteorae-server/scanners/simpleMovie"
	"github.com/meteorae/meteorae-server/scanners/stack"
	"github.com/meteorae/meteorae-server/scanners/tv"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/sdk"
)

type Scanner struct {
	Name       string
	Identifier string
	ScanFunc   func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string)
}

var scanners = map[string][]Scanner{}

func InitScannersManager() {
	scanners["internal"] = []Scanner{
		{
			Name:       audio.GetName(),
			Identifier: "tv.meteorae.scanners.audio",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				audio.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
		{
			Name:       video.GetName(),
			Identifier: "tv.meteorae.scanners.video",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				video.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
		{
			Name:       stack.GetName(),
			Identifier: "tv.meteorae.scanners.stack",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				stack.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
		{
			Name:       filter.GetName(),
			Identifier: "tv.meteorae.scanners.filter",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				filter.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
	}
	scanners["movie"] = []Scanner{
		{
			Name:       movie.GetName(),
			Identifier: "tv.meteorae.scanners.movie",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				movie.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
		{
			Name:       simpleMovie.GetName(),
			Identifier: "tv.meteorae.scanners.simpleMovie",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				simpleMovie.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
	}
	scanners["music"] = []Scanner{
		{
			Name:       music.GetName(),
			Identifier: "tv.meteorae.scanners.music",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				music.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
	}
	scanners["photo"] = []Scanner{
		{
			Name:       photos.GetName(),
			Identifier: "tv.meteorae.scanners.photo",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				photos.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
	}
	scanners["tv"] = []Scanner{
		{
			Name:       tv.GetName(),
			Identifier: "tv.meteorae.scanners.tv",
			ScanFunc: func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
				tv.Scan(path, files, dirs, mediaList, extensions, root)
			},
		},
	}
}

func GetScannerNamesForLibraryType(libraryType string) []*models.Scanner {
	names := make([]*models.Scanner, 0, len(scanners[libraryType]))

	for _, scanner := range scanners[libraryType] {
		names = append(names, &models.Scanner{
			Name:       scanner.Name,
			Identifier: scanner.Identifier,
		})
	}

	return names
}

func GetScanFuncByName(libraryType, name string) func(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	for _, scanner := range scanners[libraryType] {
		if scanner.Identifier == name {
			return scanner.ScanFunc
		}
	}

	return nil
}
