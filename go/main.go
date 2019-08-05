package main

import "gonum.org/v1/gonum/mat"

const fError = 0.001

func main() {
	tunnel := []Sphere{
		Sphere{NewVec3([]float64{0, 0, 0}), 3},
		Sphere{NewVec3([]float64{2, 0, 0}), 3},
		Sphere{NewVec3([]float64{4, 0, 0}), 3},
		Sphere{NewVec3([]float64{6, 0, 0}), 3},
	}
	disks := []Disk{
		Disk{NewVec3([]float64{1, 0, 0}), NewVec3([]float64{1, 0, 0}), 3},
		Disk{NewVec3([]float64{3, 0, 0}), NewVec3([]float64{1, 0, 0}), 3},
		Disk{NewVec3([]float64{5, 0, 0}), NewVec3([]float64{1, 0, 0}), 3},
	}
	optimizeDisks(tunnel, disks)
}

type DiskForce struct {
	positional Vec3
	torsional  mat.Dense
	radial     float64
}

func optimizeDisks(tunnel Tunnel, disks []Disk) {
	for _, disk := range disks {
		tunnel.GetMinimalDisk(disk.center, disk.normal)
	}
}
