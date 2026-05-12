package main

import (
	"math"
	"slices"
	"strings"
)

var steps = []rune("▁▂▃▄▅▆▇") // 8th rune "█" omitted to prevent gluing of rows.

func Spark(nums []float64) string {
	if len(nums) == 0 {
		return ""
	}
	indices := normalize(nums)
	var sparkline strings.Builder
	for _, index := range indices {
		sparkline.WriteRune(steps[index])
	}
	return sparkline.String()
}

func normalize(nums []float64) []int {
	total := float64(len(steps))
	lo := slices.Min(nums)
	for i := range nums {
		nums[i] -= lo
	}
	hi := slices.Max(nums)
	if hi == 0 {
		// Protect against division by zero (all values equal).
		hi = 1
	}
	indices := make([]int, len(nums))
	for i, x := range nums {
		x = (x / hi) * total
		if x == total {
			x = total - 1
		} else {
			x = math.Floor(x)
		}
		indices[i] = int(x)
	}
	return indices
}
