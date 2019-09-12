package main

import (
	"math"
	"testing"
)

func TestVec3(t *testing.T) {
	checkNaNs := func(t *testing.T, v Vec3) {
		if math.IsNaN(v.Length()) {
			t.Fatal("The vector is not supposed to contain NaNs")
		}
	}
	t.Run("Normal()", func(t *testing.T) {
		t.Run("Normal of (0, 0, 1)", func(t *testing.T) {
			n := NewVec3([]float64{0, 0, 1}).Normal()
			checkNaNs(t, n)
		})
	})
}

func TestComputeOrthogonalComplement(t *testing.T) {
	checkNaNs := func(t *testing.T, v Vec3) {
		if math.IsNaN(v.Length()) {
			t.Fatal("The Orthogonal complement vector is not supposed to contain NaNs")
		}
	}
	t.Run("OC to vector (0, 0, 1)", func(t *testing.T) {
		vecs := computeOrthogonalComplement(NewVec3([]float64{0, 0, 1}))
		checkNaNs(t, vecs[0])
		checkNaNs(t, vecs[1])
	})
}

func TestIsBasis3D(t *testing.T) {
	t.Run("Vecs (1,0,0), (0,1,0), (0,0,1)", func(t *testing.T) {
		isBasis := isBasis3D(
			NewVec3([]float64{1, 0, 0}),
			NewVec3([]float64{0, 1, 0}),
			NewVec3([]float64{0, 0, 1}),
		)
		if !isBasis {
			t.Fatal("The vectors are independed ergo isBasis3D should return true")
		}
	})
}
