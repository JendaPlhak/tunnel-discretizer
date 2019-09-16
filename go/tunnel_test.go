package main

import (
	"math"
	"testing"
)

func TestGetMinimalDisk(t *testing.T) {
	checkMinDisk := func(t *testing.T, d Disk, expCenter Vec3, expRadius float64) {
		if math.Abs(d.radius-expRadius) > fError {
			t.Errorf("Expected the min disk to have radius %f, got %f", expRadius, d.radius)
		}
		if SubVec3(d.center, expCenter).Length() > fError {
			t.Errorf("Expected the min disk to have center %v, got %v",
				expCenter, d.center)
		}
	}
	t.Run("One ball", func(t *testing.T) {
		tnl := MakeTunnel([]Sphere{
			Sphere{Vec3{0, 0, 0}, 3},
		})
		minDisk, _ := tnl.GetMinimalDisk(Vec3{0, 0, 0}, Vec3{1, 0, 0})
		checkMinDisk(t, minDisk, Vec3{0, 0, 0}, 3)
	})
	t.Run("One ball, cut on the side", func(t *testing.T) {
		tnl := MakeTunnel([]Sphere{
			Sphere{Vec3{0, 0, 0}, 3},
		})
		minDisk, _ := tnl.GetMinimalDisk(Vec3{1, 0, 0}, Vec3{1, 0, 0})
		checkMinDisk(t, minDisk, Vec3{1, 0, 0}, 2.828427)
	})
	t.Run("Two balls, cut in the middle", func(t *testing.T) {
		tnl := MakeTunnel([]Sphere{
			Sphere{Vec3{0, 0, 0}, 3},
			Sphere{Vec3{2, 0, 0}, 3},
		})
		minDisk, _ := tnl.GetMinimalDisk(Vec3{1, 0, 0}, Vec3{1, 0, 0})
		checkMinDisk(t, minDisk, Vec3{1, 0, 0}, 2.828427)
	})
	t.Run("Three balls, sharp turn", func(t *testing.T) {
		tnl := MakeTunnel([]Sphere{
			Sphere{Vec3{0, 0, 0}, 1},
			Sphere{Vec3{1, 0, 0}, 1},
		})
		minDisk, _ := tnl.GetMinimalDisk(Vec3{1, 0, 0}, Vec3{1, 1, 0})

		r := (1 + math.Sqrt(2)) / 2
		x := r / math.Sqrt(2)
		y := 1 - x
		checkMinDisk(t, minDisk, Vec3{x, y, 0}, r)
	})
}
