package main

import "math/rand"

func RandIntAt(min, max int64) int64 {

	if min > max || min == 0 || max == 0 {
		return max
	}

	return rand.Int63n(max-min) + min
}
