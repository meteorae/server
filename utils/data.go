package utils

import (
	"runtime"
	"sort"
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

func ReverseSlice(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

func RankMapStringInt(values map[string]int) []string {
	type kv struct {
		Key   string
		Value int
	}

	ss := make([]kv, 0, len(values))

	for k, v := range values {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	ranked := make([]string, len(values))

	for i, kv := range ss {
		ranked[i] = kv.Key
	}

	return ranked
}
