package utils

import (
	"regexp"
	"strings"
)

func CleanSortTitle(str string) string {
	// This Regex should cover most initial articles for common languages
	// See: https://web.library.yale.edu/cataloging/music/initial-articles-listed-by-language
	// See: https://library.princeton.edu/departments/tsd/katmandu/catcopy/article.html
	// See: https://www.loc.gov/marc/bibliographic/bdapp-e.html
	reg := regexp.MustCompile(`(?i)^(?:a |an |d'|de |the |ye |l'|la |le |les |un |une |el |els |en |una |de |den |det |en |et |het |een |eene |'n |'t |das |dem |den |der |des |die |ein |eine |einem |einen |einer |eines |'s |gl'|gli |i |il |lo |un'|uno |las |lo |los |o |unei |unui |dei |ei |eit |ett |y | yr |)(.*)`)
	res := reg.ReplaceAllString(str, "${1}")

	return res
}

func RemoveUnwantedCharacters(str string) string {
	reg := regexp.MustCompile(`(_|,|\.|\(|\)|\[|\]|-|:)`)
	res := reg.ReplaceAllString(str, " ")

	return strings.TrimSpace(res)
}
