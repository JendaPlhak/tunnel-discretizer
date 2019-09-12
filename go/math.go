package main

import (
	"math"
	"math/rand"
)

const fError = 0.001

func MaxInt(x, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}

func MinInt(x, y int) int {
	if x > y {
		return y
	} else {
		return x
	}
}

func RandFloat64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func LogWithBase(base, x float64) float64 {
	return math.Log(x) / math.Log(base)
}