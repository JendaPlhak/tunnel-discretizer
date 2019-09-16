package main

import (
	"testing"
)

func TestOptimizerCompleteTunnel(t *testing.T) {
	t.Run("../tun_cl_002_1.pdb", func(t *testing.T) {
		tunnel := LoadTunnelFromPdbFile("../tun_cl_002_1.pdb")
		tunnel.Spheres = tunnel.Spheres[:3]
		disks := generateInitialDisks(tunnel)
		optimizeDisks(tunnel, disks, 1000)
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
		optimizeDisks(tunnel, disks, 5)
		if SubVec3(disks[1].normal, Vec3{1, 0, 0}).Length() > 0.001 {
			t.Error("Optimizer should shift the middle disk towards normal (1,0,0)")
		}
	})
}
