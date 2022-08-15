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
	// TODO: Split this per language, so it's easier to read. Also make it configurable.
	reg := regexp.MustCompile(`(?i)^(?:a |an |d'|de |the |ye |l'|la |le |les |un |une |el |els |en |una |de |den |det |en |et |het |een |eene |'n |'t |das |dem |den |der |des |die |ein |eine |einem |einen |einer |eines |'s |gl'|gli |i |il |lo |un'|uno |las |lo |los |o |unei |unui |dei |ei |eit |ett |y | yr |)(.*)`) //nolint: lll
	res := reg.ReplaceAllString(str, "${1}")

	return res
}

func RemoveUnwantedCharacters(str string) string {
	reg := regexp.MustCompile(`(_|,|\.|\(|\)|\[|\]|-|:)`)
	result := reg.ReplaceAllString(str, " ")

	// Split on " Aka "
	result = strings.Split(result, " Aka ")[0]

	return strings.TrimSpace(result)
}

func FindNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}

	return results
}
