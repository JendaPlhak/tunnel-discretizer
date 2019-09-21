package main

import (
	"math"
	"testing"
)

func TestOptimizerCompleteTunnel(t *testing.T) {
	t.Run("../tun_cl_002_1.pdb", func(t *testing.T) {
		tunnel := LoadTunnelFromPdbFile("../tun_cl_002_1.pdb")
		tunnel.Spheres = tunnel.Spheres[:3]
		disks := generateInitialDisks(tunnel)
		OptimizeDisks(tunnel, disks)
	})
}

func TestOptimizerElementaryCases(t *testing.T) {
	getCheckedMinDisk := func(t *testing.T, tunnel Tunnel, point, normal Vec3) Disk {
		disk, ok := tunnel.GetMinimalDisk(point, normal)
		if !ok {
			t.Error("Failed to generate initial disk.")
		}
		return disk
	}
	t.Run("Two balls, two disks, both on the second ball", func(t *testing.T) {
		tunnel := MakeTunnel([]Sphere{
			Sphere{Vec3{-2, 0, 0}, 1.1},
			Sphere{Vec3{0, 0, 0}, 1},
		})
		disks := []Disk{
			getCheckedMinDisk(t, tunnel, Vec3{-0.5, 0, 0}, Vec3{1, 0, 0}),
			Disk{center: Vec3{0, 0, 0}, normal: Vec3{1, -1, 0}.Normalized(), radius: 1},
			getCheckedMinDisk(t, tunnel, Vec3{0.5, 0, 0}, Vec3{1, 0, 0}),
		}
		OptimizeDisks(tunnel, disks)
		if SubVec3(disks[1].normal, Vec3{1, 0, 0}).Length() > 0.001 {
			t.Error("Optimizer should shift the middle disk towards normal (1,0,0)")
		}
	})
}

func TestOrientedDisk(t *testing.T) {
	t.Run("Test directions", func(t *testing.T) {
		testEqual := func(exp, v Vec3) {
			if SubVec3(v, exp).Length() > fError {
				t.Errorf("Expected vector %v, got %v", exp, v)
			}
		}
		disk := OrientedDisk{
			Disk: Disk{
				center: Vec3{0, 0, 0},
				normal: Vec3{1, 0, 0},
				radius: 1,
			},
			up: Vec3{0, 1, 0},
		}
		testEqual(Vec3{0, 1, 0}, disk.getUp())
		testEqual(Vec3{0, -1, 0}, disk.getDown())
		testEqual(Vec3{0, 0, 1}, disk.getRight())
		testEqual(Vec3{0, 0, -1}, disk.getLeft())
	})
	t.Run("Test getPointQuadrant", func(t *testing.T) {
		disk := OrientedDisk{
			Disk: Disk{
				center: Vec3{0, 0, 0},
				normal: Vec3{1, 0, 0},
				radius: 1,
			},
			up: Vec3{0, 1, 0},
		}
		if idx := disk.getPrecedingDir(Vec3{0, 0.5, 0.5}); idx != UpDirIdx {
			t.Errorf("Expected the first Dir, got %v", idx)
		}
		if idx := disk.getPrecedingDir(Vec3{0, -0.5, 0.5}); idx != RightDirIdx {
			t.Errorf("Expected the second Dir, got %v", idx)
		}
		if idx := disk.getPrecedingDir(Vec3{0, -0.5, -0.5}); idx != DownDirIdx {
			t.Errorf("Expected the third Dir, got %v", idx)
		}
		if idx := disk.getPrecedingDir(Vec3{0, 0.5, -0.5}); idx != LeftDirIdx {
			t.Errorf("Expected the fourth Dir, got %v", idx)
		}
	})
	t.Run("Test getForceField", func(t *testing.T) {
		checkBalance := func(t *testing.T, ff ForceField) {
			if AddVec3(ff.Top, ff.Bottom).Length() > fError {
				t.Error("Force vectors Top and Bottom are imbalanced")
			}
			if AddVec3(ff.Right, ff.Left).Length() > fError {
				t.Error("Force vectors Right and Left are imbalanced")
			}
		}
		compareForceFields := func(t *testing.T, ff ForceField, ffExp ForceField) {
			if !ffExp.Top.Equals(ff.Top) {
				t.Errorf("Expected the Top force to be %v, got %v", ffExp.Top, ff.Top)
			}
			if !ffExp.Right.Equals(ff.Right) {
				t.Errorf("Expected the Right force to be %v, got %v", ffExp.Right, ff.Right)
			}
			if !ffExp.Bottom.Equals(ff.Bottom) {
				t.Errorf("Expected the Bottom force to be %v, got %v", ffExp.Bottom, ff.Bottom)
			}
			if !ffExp.Left.Equals(ff.Left) {
				t.Errorf("Expected the Left force to be %v, got %v", ffExp.Left, ff.Left)
			}
		}
		disk := OrientedDisk{
			Disk: Disk{
				center: Vec3{0, 0, 0},
				normal: Vec3{1, 0, 0},
				radius: 1,
			},
			up: Vec3{0, 1, 0},
		}
		t.Run("force aligned with up", func(t *testing.T) {
			ff := disk.getForceField(Vec3{0, 1, 0}, Vec3{2, 0, 0})
			checkBalance(t, ff)
			compareForceFields(t, ff, ForceField{
				Top:    Vec3{2, 0, 0},
				Right:  Vec3{0, 0, 0},
				Bottom: Vec3{-2, 0, 0},
				Left:   Vec3{0, 0, 0},
			})
		})
		t.Run("force between Top and Right", func(t *testing.T) {
			ff := disk.getForceField(
				Vec3{0, 0.5, 0.5},
				Vec3{2, 0, 0},
			)
			checkBalance(t, ff)
			compareForceFields(t, ff, ForceField{
				Top:    Vec3{1, 0, 0},
				Right:  Vec3{1, 0, 0},
				Bottom: Vec3{-1, 0, 0},
				Left:   Vec3{-1, 0, 0},
			})
		})
		t.Run("force aligned with Top, skewed up", func(t *testing.T) {
			ff := disk.getForceField(
				Vec3{0, 1, 0},
				Vec3{0.5, 0.5, 0},
			)
			checkBalance(t, ff)
			compareForceFields(t, ff, ForceField{
				Top:    Vec3{math.Cos(math.Pi / 4), 0, 0},
				Right:  Vec3{0, 0, 0},
				Bottom: Vec3{-math.Cos(math.Pi / 4), 0, 0},
				Left:   Vec3{0, 0, 0},
			})
		})
	})

}
