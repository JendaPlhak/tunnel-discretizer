package main

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
