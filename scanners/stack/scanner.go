package stack

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/meteorae/go-difflib/difflib"
	"github.com/meteorae/meteorae-server/models"
	"github.com/meteorae/meteorae-server/scanners/video"
)

func GetName() string {
	return "Media Stack Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]models.Item, extensions []string, root string) {
	var stackMap map[string][]models.Item

	stackDiffs := "123456789abcdefghijklmn"
	stackSuffixes := []string{
		"cd",
		"dvd",
		"part",
		"pt",
		"disk",
		"disc",
		"scene",
	}

	sort.Slice(*mediaList, func(i, j int) bool {
		mediaI, ok := (*mediaList)[i].(*models.Movie)
		mediaJ, ok2 := (*mediaList)[j].(*models.Movie)
		if ok && ok2 {
			return sort.StringsAreSorted([]string{mediaI.Parts[0].FilePath, mediaJ.Parts[0].FilePath})
		}

		return false
	})

	count := 0
	for _, media := range *mediaList {
		m1 := (*mediaList)[count]
		m2 := (*mediaList)[count+1]
		media1, ok := m1.(*models.Movie)
		media2, ok2 := m2.(*models.Movie)

		if ok && ok2 {
			f1 := filepath.Base(media1.Parts[0].FilePath)
			f2 := filepath.Base(media2.Parts[0].FilePath)

			// This uses SequenceMatcher to find how many differences are in the two paths.
			// If there is only one replaced character and that character is in stackDiffs,
			// we attempt to process it as a stack of files.
			opcodes := difflib.NewMatcher(splitChars(f1), splitChars(f2)).GetOpCodes()
			if len(opcodes) == 3 {
				if string(opcodes[1].Tag) == "replace" {
					if (opcodes[1].I2-opcodes[1].I1 == 1) && opcodes[1].J2-opcodes[1].J1 == 1 {
						character := strings.ToLower(f1[opcodes[1].I1:opcodes[1].I2])
						if strings.Contains(stackDiffs, character) {
							root := f1[:opcodes[1].I1]

							// Handle the X of Y cases
							xOfy := false
							if strings.HasPrefix(strings.Trim(strings.ToLower(f1[opcodes[1].I1+1:]), " "), "of") {
								xOfy = true
							}

							// Remove leading zeroes from part numbers.
							if root[len(root)-1:] == "0" {
								root = root[:len(root)-1]
							}

							// Properly handle stuff like Kill Bill Vol. 1 and Vol. 2
							if !(strings.HasSuffix(strings.Trim((strings.ToLower(root)), " "), "vol")) &&
								!(strings.HasSuffix(strings.Trim((strings.ToLower(root)), " "), "volume")) {
								foundSuffix := false

								for _, suffix := range stackSuffixes {
									if strings.HasSuffix(strings.Trim((strings.ToLower(root)), " "), suffix) {
										root = root[:len(root)-len(suffix)]
										foundSuffix = true

										break
									}

									if foundSuffix && xOfy {
										// In this case, the name probably had a suffix, so replace it
										name, _ := video.CleanName(root)

										// Assume this is a movie, since it's a video anyway
										if mediaMovie, ok := media.(models.Movie); ok {
											mediaMovie.Title = name

											media = mediaMovie
										}

										if _, ok := stackMap[root]; ok {
											stackMap[root] = append(stackMap[root], m2)
										} else {
											stackMap[root] = []models.Item{m1}
											stackMap[root] = append(stackMap[root], m2)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for stack := range stackMap {
		for mediaIndex, media := range stackMap[stack][1:] {
			mediaMetadata, ok := media.(models.MetadataModel)
			mediaMetadataSlice, ok2 := stackMap[stack][0].(models.MetadataModel)

			if ok && ok2 {
				mediaMetadataSlice.Parts = append(mediaMetadataSlice.Parts, mediaMetadata.Parts...)

				stackMap[stack][0] = mediaMetadataSlice
			}

			*mediaList = append((*mediaList)[:mediaIndex], (*mediaList)[mediaIndex+1:]...)
		}
	}
}

func splitChars(s string) []string {
	chars := make([]string, 0, len(s))
	// Assume ASCII inputs
	for i := 0; i != len(s); i++ {
		chars = append(chars, string(s[i]))
	}

	return chars
}
