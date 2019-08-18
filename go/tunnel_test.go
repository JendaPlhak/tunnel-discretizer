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
				expCenter.RawVector(), d.center.RawVector())
		}
	}
	t.Run("One ball", func(t *testing.T) {
		tnl := Tunnel{
			Sphere{NewVec3([]float64{0, 0, 0}), 3},
		}
		minDisk := tnl.GetMinimalDisk(NewVec3([]float64{0, 0, 0}), NewVec3([]float64{1, 0, 0}))
		checkMinDisk(t, minDisk, NewVec3([]float64{0, 0, 0}), 3)
	})
	t.Run("Two balls, cut in the middle", func(t *testing.T) {
		tnl := Tunnel{
			Sphere{NewVec3([]float64{0, 0, 0}), 3},
			Sphere{NewVec3([]float64{2, 0, 0}), 3},
		}
		minDisk := tnl.GetMinimalDisk(NewVec3([]float64{1, 0, 0}), NewVec3([]float64{1, 0, 0}))
		checkMinDisk(t, minDisk, NewVec3([]float64{1, 0, 0}), 2.828427)
	})
}
