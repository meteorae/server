package stack

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/ianbruene/go-difflib/difflib"
	"github.com/meteorae/meteorae-server/database"
	"github.com/meteorae/meteorae-server/scanners/video"
)

func GetName() string {
	return "Media Stack Scanner"
}

func Scan(path string, files, dirs *[]string, mediaList *[]database.ItemMetadata, extensions []string, root string) {
	var stackMap map[string][]database.ItemMetadata

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
		return sort.StringsAreSorted([]string{(*mediaList)[i].MediaParts[0].FilePath, (*mediaList)[j].MediaParts[0].FilePath})
	})

	count := 0
	for _, media := range *mediaList {
		m1 := (*mediaList)[count]
		m2 := (*mediaList)[count+1]
		f1 := filepath.Base(m1.MediaParts[0].FilePath)
		f2 := filepath.Base(m2.MediaParts[0].FilePath)

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
									media.Title = name

									if _, ok := stackMap[root]; ok {
										stackMap[root] = append(stackMap[root], m2)
									} else {
										stackMap[root] = []database.ItemMetadata{m1}
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

	for stack := range stackMap {
		for mediaIndex, media := range stackMap[stack][1:] {
			stackMap[stack][0].MediaParts = append(stackMap[stack][0].MediaParts, media.MediaParts...)
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
