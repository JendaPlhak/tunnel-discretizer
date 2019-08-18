package main

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

const fError = 0.001
const Delta = 0.5
const Eps = Delta / 10.

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
	for i := 1; i < len(disks)-1; i++ {
		left, middle, right := disks[i-1], disks[i], disks[i+1]
		disks[i] = optimizeDiskLocally(tunnel, left, middle, right)
		fmt.Println(disks[i])
	}
}

func optimizeDiskLocally(tunnel Tunnel, left, middle, right Disk) Disk {
	minDisk := tunnel.GetMinimalDisk(middle.center, middle.normal)

	movedLeft := moveTowardsByEps(middle, left)
	movedToMin := moveTowardsByEps(middle, minDisk)
	movedRight := moveTowardsByEps(middle, right)

	energyToTheLeft := evaluateEnergy(tunnel, left, movedLeft, right)
	energyToTheMin := evaluateEnergy(tunnel, left, movedToMin, right)
	energyToTheRight := evaluateEnergy(tunnel, left, movedRight, right)
	fmt.Println(energyToTheLeft, energyToTheMin, energyToTheRight)

	energies := []DiskWithEnergy{
		DiskWithEnergy{energyToTheLeft, movedLeft},
		DiskWithEnergy{energyToTheMin, movedToMin},
		DiskWithEnergy{energyToTheRight, movedRight},
	}
	return getMinEnergyDisk(energies)
}

type DiskWithEnergy struct {
	energy float64
	disk   Disk
}

func getMinEnergyDisk(energies []DiskWithEnergy) Disk {
	minDiskWithEnergy := energies[0]
	for i := 1; i < len(energies); i++ {
		if energies[i].energy < minDiskWithEnergy.energy {
			minDiskWithEnergy = energies[i]
		}
	}
	return minDiskWithEnergy.disk
}

func moveTowardsByEps(source, target Disk) Disk {
	l := SubVec3(source.center, target.center).Length()
	alpha := math.Min(Eps, l) / l
	return disksLinearCombination(source, 1-alpha, target, alpha)
}

func evaluateEnergy(tunnel Tunnel, left, middle, right Disk) float64 {
	return evalEnergyTowardsMindisk(tunnel, middle)
}

func evalEnergyTowardsMindisk(tunnel Tunnel, middle Disk) float64 {
	minDisk := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if minDisk.radius > middle.radius {
		return math.Pow(minDisk.radius-middle.radius, 3)
	} else if middle.radius-minDisk.radius < Delta*0.5 {
		return 0
	} else {
		return math.Pow(middle.radius-minDisk.radius, 2)
	}
}
