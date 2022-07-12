package utils

import (
	"runtime"
	"time"

	"github.com/rs/zerolog/log"
)

func IsStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func RemoveStringFromSlice(a string, list []string) []string {
	for i, b := range list {
		if b == a {
			list = append(list[:i], list[i+1:]...)
		}
	}

	return list
}

// RemoveDuplicatesFromSlice removes duplicates from a slice of any types implementing the comparable interface.
func RemoveDuplicatesFromSlice[T comparable](s []T) []T {
	inResult := make(map[T]bool)

	var result []T

	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true

			result = append(result, str)
		}
	}

	return result
}

func TimeTrack(start time.Time) {
	elapsed := time.Since(start)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1) // nolint:dogsled

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	log.Debug().Str("function", funcObj.Name()).Msgf("Took %s", elapsed)
}
