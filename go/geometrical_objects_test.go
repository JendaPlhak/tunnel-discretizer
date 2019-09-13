package main

import (
	"math"
	"testing"
)

func TestSphere(t *testing.T) {
	checkVecEqual := func(t *testing.T, u, v Vec3) {
		if math.Abs(SubVec3(u, v).Length()) > fError {
			t.Errorf("Expected vectors to be the same, %v vs %v", u.RawVector(), v.RawVector())
		}
	}
	t.Run("intersectionLine", func(t *testing.T) {
		s := Sphere{NewVec3([]float64{0, 0, 0}), 3}
		l := Line{point: NewVec3([]float64{0, 0, 0}), dir: NewVec3([]float64{0, 1, 0})}
		points := s.intersectionLine(l)
		checkVecEqual(t, points[0], NewVec3([]float64{0, 3, 0}))
		checkVecEqual(t, points[1], NewVec3([]float64{0, -3, 0}))
	})
}

func TestPlane(t *testing.T) {
	checkVecEqual := func(t *testing.T, u, expU Vec3) {
		if math.Abs(SubVec3(u, expU).Length()) > fError {
			t.Errorf("Expected vector to be %v, got %v", expU.RawVector(), u.RawVector())
		}
	}
	t.Run("getBaseVectors", func(t *testing.T) {
		t.Run("zero center, normal [0,0,1]", func(t *testing.T) {
			p := MakePlane(
				NewVec3([]float64{0, 0, 0}),
				NewVec3([]float64{0, 0, 1}),
			)
			v1, v2 := p.getBaseVectors()
			if math.IsNaN(v1.Length()) {
				t.Fatal("The base vector is not supposed to contain NaNs")
			} else if math.IsNaN(v2.Length()) {
				t.Fatal("The base vector is not supposed to contain NaNs")
			} else if math.Abs(v1.Length()-1) > fError {
				t.Fatal("The base vector is supposed to be normalized.")
			} else if math.Abs(v2.Length()-1) > fError {
				t.Fatal("The base vector is supposed to be normalized.")
			}
		})
	})
	t.Run("intersectionWithSphere", func(t *testing.T) {
		t.Run("center at [0,0,0]", func(t *testing.T) {
			s := Sphere{NewVec3([]float64{0, 0, 0}), 3}
			p := MakePlane(
				NewVec3([]float64{0, 0, 0}),
				NewVec3([]float64{1, 0, 0}),
			)
			circle, ok := p.intersectionWithSphere(s)
			if !ok {
				t.Fatal("The sphere and plane unexpectedly don't intersect.")
			}
			if circle.radius != 3 {
				t.Errorf("Expected cut radius 3 got %f", circle.radius)
			}
			center3D := p.transformPointTo3D(circle.center)
			checkVecEqual(t, center3D, NewVec3([]float64{0, 0, 0}))
		})
		t.Run("center at [1,0,0]", func(t *testing.T) {
			s := Sphere{NewVec3([]float64{0, 0, 0}), 3}
			p := MakePlane(
				NewVec3([]float64{1, 0, 0}),
				NewVec3([]float64{1, 0, 0}),
			)
			circle, ok := p.intersectionWithSphere(s)
			if !ok {
				t.Fatal("The sphere and plane unexpectedly don't intersect.")
			}
			if circle.radius-2.828427 > fError {
				t.Errorf("Expected cut radius 2.828427 got %f", circle.radius)
			}
			center3D := p.transformPointTo3D(circle.center)
			checkVecEqual(t, center3D, NewVec3([]float64{1, 0, 0}))
		})
	})
	t.Run("intersectionWithLine", func(t *testing.T) {
		t.Run("orthogonal line, inter at [0,0,0]", func(t *testing.T) {
			l := Line{point: NewVec3([]float64{0, 0, 0}), dir: NewVec3([]float64{0, 0, 1})}
			p := MakePlane(
				NewVec3([]float64{0, 0, 0}),
				NewVec3([]float64{0, 0, 1}),
			)
			inter, ok := p.intersectionWithLine(l)
			if !ok {
				t.Fatal("The line and plane unexpectedly don't intersect.")
			}
			checkVecEqual(t, inter, NewVec3([]float64{0, 0, 0}))
		})
		t.Run("orthogonal line, inter at [1,0,0]", func(t *testing.T) {
			l := Line{point: NewVec3([]float64{1, 0, 5}), dir: NewVec3([]float64{0, 0, 1})}
			p := MakePlane(
				NewVec3([]float64{0, 0, 0}),
				NewVec3([]float64{0, 0, 1}),
			)
			inter, ok := p.intersectionWithLine(l)
			if !ok {
				t.Fatal("The line and plane unexpectedly don't intersect.")
			}
			checkVecEqual(t, inter, NewVec3([]float64{1, 0, 0}))
		})
	})
}

func TestGetDisksDistances(t *testing.T) {
	checkDistances := func(t *testing.T, l1, l2, exp1, exp2 float64) {
		if math.Abs(l1-exp1) > fError {
			t.Errorf("Expected the first distance to be %f, got %f", exp1, l1)
		}
		if math.Abs(l2-exp2) > fError {
			t.Errorf("Expected the second distance to be %f, got %f", exp2, l2)
		}
	}

	t.Run("two parallel disks", func(t *testing.T) {
		t.Run("same radius", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 3}
			d2 := Disk{NewVec3([]float64{2, 0, 0}), NewVec3([]float64{1, 0, 0}), 3}

			l1, l2 := GetDisksDistances(d1, d2)
			checkDistances(t, l1, l2, 2, 2)
		})
		t.Run("different radius", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 1}
			d2 := Disk{NewVec3([]float64{3, 0, 0}), NewVec3([]float64{1, 0, 0}), 5}

			l1, l2 := GetDisksDistances(d1, d2)
			checkDistances(t, l1, l2, 5, 5)
		})
		t.Run("negative distance", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{1, 0, 0}), NewVec3([]float64{1, 0, 0}), 2}
			d2 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 2}

			l1, l2 := GetDisksDistances(d1, d2)
			checkDistances(t, l1, l2, -1, -1)
		})
		t.Run("different radius, negative distance", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{3, 0, 0}), NewVec3([]float64{1, 0, 0}), 1}
			d2 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 5}

			l1, l2 := GetDisksDistances(d1, d2)
			checkDistances(t, l1, l2, -5, -5)
		})
	})
	t.Run("two non-parallel disks", func(t *testing.T) {
		t.Run("right triangle", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 3}
			d2 := Disk{NewVec3([]float64{4, 0, 0}), NewVec3([]float64{3, 4, 0}).Normalized(), 5}

			l1, l2 := GetDisksDistances(d1, d2)
			checkDistances(t, l1, l2, 0, 8)
		})
		t.Run("right triangle. negative distance", func(t *testing.T) {
			d1 := Disk{NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}), 3}
			d2 := Disk{NewVec3([]float64{4, 0, 0}), NewVec3([]float64{3, 4, 0}).Normalized(), 5}

			l1, l2 := GetDisksDistances(d2, d1)
			checkDistances(t, l1, l2, -8, 0)
		})
	})
}
