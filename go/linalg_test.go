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
			n := Vec3{0, 0, 1}.Normal()
			checkNaNs(t, n)
		})
	})
	t.Run("IsLinearCombinationOf()", func(t *testing.T) {
		t.Run("(2, 0, 0) and (4, 0, 0)", func(t *testing.T) {
			if !(Vec3{2, 0, 0}.IsLinearCombinationOf(Vec3{4, 0, 0})) {
				t.Fatal("The vectors should be evaluated as linearly dependent")
			}
		})
		t.Run("(2, 0, 0) and (4, 0, 0)", func(t *testing.T) {
			if !(Vec3{-2, 0, 0}.IsLinearCombinationOf(Vec3{4, 0, 0})) {
				t.Fatal("The vectors should be evaluated as linearly dependent")
			}
		})
		t.Run("(2, 0, 0) and (4, 0, 0)", func(t *testing.T) {
			if (Vec3{2, 0, 0}.IsLinearCombinationOf(Vec3{4, 0, 1})) {
				t.Fatal("The vectors should be evaluated as linearly independent")
			}
		})
	})
}

func TestMat3x3(t *testing.T) {
	checkEqual := func(t *testing.T, A, B Mat3x3) {
		if A.a00 != B.a00 {
			t.Errorf("Unexpected matrix value at position 00. Expected %f, got %f", A.a00, B.a00)
		} else if A.a01 != B.a01 {
			t.Errorf("Unexpected matrix value at position 01. Expected %f, got %f", A.a01, B.a01)
		} else if A.a02 != B.a02 {
			t.Errorf("Unexpected matrix value at position 02. Expected %f, got %f", A.a02, B.a02)
		} else if A.a10 != B.a10 {
			t.Errorf("Unexpected matrix value at position 10. Expected %f, got %f", A.a10, B.a10)
		} else if A.a11 != B.a11 {
			t.Errorf("Unexpected matrix value at position 11. Expected %f, got %f", A.a11, B.a11)
		} else if A.a12 != B.a12 {
			t.Errorf("Unexpected matrix value at position 12. Expected %f, got %f", A.a12, B.a12)
		} else if A.a20 != B.a20 {
			t.Errorf("Unexpected matrix value at position 20. Expected %f, got %f", A.a20, B.a20)
		} else if A.a21 != B.a21 {
			t.Errorf("Unexpected matrix value at position 21. Expected %f, got %f", A.a21, B.a21)
		} else if A.a22 != B.a22 {
			t.Errorf("Unexpected matrix value at position 22. Expected %f, got %f", A.a22, B.a22)
		}
	}
	t.Run("Test Det()", func(t *testing.T) {
		t.Run("diagonal matrix", func(t *testing.T) {
			A := Mat3x3{
				1, 0, 0,
				0, 1, 0,
				0, 0, -1,
			}
			if det := A.Det(); det != -1 {
				t.Errorf("Expected the determinant to be -1, got %f", det)
			}
		})
	})
	t.Run("Test Scale()", func(t *testing.T) {
		t.Run("Scale by -1", func(t *testing.T) {
			A := Mat3x3{
				1, -2, 3,
				-4, 5, -6,
				7, -8, 9,
			}
			ExpScale := Mat3x3{
				-1, 2, -3,
				4, -5, 6,
				-7, 8, -9,
			}
			A.Scale(-1)
			checkEqual(t, ExpScale, A)
		})
	})
	t.Run("Test Inversed()", func(t *testing.T) {
		t.Run("diagonal matrix", func(t *testing.T) {
			A := Mat3x3{
				1, 0, 0,
				0, 1, 0,
				0, 0, -1,
			}
			ExpInv := Mat3x3{
				1, 0, 0,
				0, 1, 0,
				0, 0, -1,
			}
			checkEqual(t, ExpInv, A.Inversed())
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
		vecs := computeOrthogonalComplement(Vec3{0, 0, 1})
		checkNaNs(t, vecs[0])
		checkNaNs(t, vecs[1])
	})
}

func TestIsBasis3D(t *testing.T) {
	t.Run("Vecs (1,0,0), (0,1,0), (0,0,1)", func(t *testing.T) {
		isBasis := isBasis3D(
			Vec3{1, 0, 0},
			Vec3{0, 1, 0},
			Vec3{0, 0, 1},
		)
		if !isBasis {
			t.Fatal("The vectors are independed ergo isBasis3D should return true")
		}
	})
}

func TestRotAngle3(t *testing.T) {
	checkAngle := func(t *testing.T, exp, alpha float64) {
		if math.Abs(exp-alpha) > fError {
			t.Errorf("Expected the rotation angle to be %f, got %f", exp, alpha)
		}
	}
	t.Run("OC to vector (0, 0, 1)", func(t *testing.T) {
		axis := Vec3{1, 0, 0}
		v1 := Vec3{0, 1, 0}
		checkAngle(t, math.Pi/4, RotAngle3(axis, v1, Vec3{0, 0.5, 0.5}))
		checkAngle(t, math.Pi/2, RotAngle3(axis, v1, Vec3{0, 0, 1}))
		checkAngle(t, 3/4.*math.Pi, RotAngle3(axis, v1, Vec3{0, -0.5, 0.5}))
		checkAngle(t, math.Pi, RotAngle3(axis, v1, Vec3{0, -1, 0}))
		checkAngle(t, 5/4.*math.Pi, RotAngle3(axis, v1, Vec3{0, -0.5, -0.5}))
		checkAngle(t, 3/2.*math.Pi, RotAngle3(axis, v1, Vec3{0, 0, -1}))
		checkAngle(t, 7/4.*math.Pi, RotAngle3(axis, v1, Vec3{0, 0.5, -0.5}))
	})
}
