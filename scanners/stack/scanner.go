package stack

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/meteorae/go-difflib/difflib"
	"github.com/meteorae/meteorae-server/scanners/video"
	"github.com/meteorae/meteorae-server/sdk"
	"github.com/rs/zerolog/log"
)

func GetName() string {
	return "Media Stack Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]sdk.Item, extensions []string, root string) {
	log.Debug().Str("scanner", GetName()).Msgf("Scanning %s", path)

	var stackMap map[string][]sdk.Item

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
		mediaI, ok := (*mediaList)[i].(*sdk.ItemInfo)
		mediaJ, ok2 := (*mediaList)[j].(*sdk.ItemInfo)
		if ok && ok2 {
			return sort.StringsAreSorted([]string{mediaI.Parts[0], mediaJ.Parts[0]})
		}

		return false
	})

	count := 0
	for _, media := range *mediaList {
		m1 := (*mediaList)[count]

		if count+1 < len(*mediaList) {
			m2 := (*mediaList)[count+1]
			media1, ok := m1.(*sdk.ItemInfo)
			media2, ok2 := m2.(*sdk.ItemInfo)

			if ok && ok2 {
				f1 := filepath.Base(media1.Parts[0])
				f2 := filepath.Base(media2.Parts[0])

				// This uses SequenceMatcher to find how many differences are in the two paths.
				// If there is only one replaced character and that character is in stackDiffs,
				// we attempt to process it as a stack of files.
				opcodes := difflib.NewMatcher(splitChars(f1), splitChars(f2)).GetOpCodes()
				if len(opcodes) == 3 { //nolint:gomnd
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

											if mediaMovie, ok := media.(sdk.ItemInfo); ok {
												mediaMovie.Title = name

												media = mediaMovie
											}

											if _, ok := stackMap[root]; ok {
												stackMap[root] = append(stackMap[root], m2)
											} else {
												stackMap[root] = []sdk.Item{m1}
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
	}

	for stack := range stackMap {
		for mediaIndex, media := range stackMap[stack][1:] {
			mediaMetadata, ok := media.(sdk.ItemInfo)
			mediaMetadataSlice, ok2 := stackMap[stack][0].(sdk.ItemInfo)

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
