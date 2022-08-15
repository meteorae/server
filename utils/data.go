package utils

import (
	"sort"
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

	stringSliceMap := make([]kv, 0, len(values))

	for k, v := range values {
		stringSliceMap = append(stringSliceMap, kv{k, v})
	}

	sort.Slice(stringSliceMap, func(i, j int) bool {
		return stringSliceMap[i].Value > stringSliceMap[j].Value
	})

	ranked := make([]string, 0, len(values))

	for i, kv := range stringSliceMap {
		ranked[i] = kv.Key
	}

	return ranked
}
