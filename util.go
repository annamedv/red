package main

import (
	"maps"
	"slices"
)

func mapKeys(m map[string]interface{}) []string {
	return slices.Sorted(maps.Keys(m))
}
