package main

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
)

type DiskForce struct {
	positional Vec3
	torsional  mat.Dense
	radial     float64
}

func optimizeDisks(tunnel Tunnel, disks []Disk) {
	totalEnergy := getTotalEnergy(tunnel, disks)
	for round := 0; round < 50; round++ {
		fmt.Printf("Beginning round %d, Total energy: %f\n", round, totalEnergy)

		for i := 1; i < len(disks)-1; i++ {
			left, middle, right := disks[i-1], disks[i], disks[i+1]
			disks[i] = optimizeDiskLocally(tunnel, left, middle, right)
		}
		newTotalEnergy := getTotalEnergy(tunnel, disks)
		if totalEnergy-newTotalEnergy < 0.1 {
			fmt.Printf("Energy stabilized at %f, ending optimization\n", newTotalEnergy)
			break
		}
		totalEnergy = newTotalEnergy
	}

}

func getTotalEnergy(tunnel Tunnel, disks []Disk) float64 {
	totalEnergy := 0.
	for i := 1; i < len(disks)-1; i++ {
		totalEnergy += evaluateEnergy(tunnel, disks[i-1], disks[i], disks[i+1])
	}
	return totalEnergy
}

func optimizeDiskLocally(tunnel Tunnel, left, middle, right Disk) Disk {
	minDisk := tunnel.GetMinimalDisk(middle.center, middle.normal)

	movedLeft := moveTowardsByEps(middle, left)
	movedToMin := moveTowardsByEps(middle, minDisk)
	movedRight := moveTowardsByEps(middle, right)

	energyToTheLeft := evaluateEnergy(tunnel, left, movedLeft, right)
	energyToTheMin := evaluateEnergy(tunnel, left, movedToMin, right)
	energyToTheRight := evaluateEnergy(tunnel, left, movedRight, right)
	energyDoNothing := evaluateEnergy(tunnel, left, middle, right)

	energies := []DiskWithEnergy{
		DiskWithEnergy{energyDoNothing, middle},
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
	l1 := SubVec3(source.center, target.center).Length()
	l2 := math.Abs(source.radius - target.radius)
	l := math.Max(l1, l2)
	if l == 0. {
		return source
	}
	alpha := math.Min(Eps, l) / l
	return disksLinearCombination(source, 1-alpha, target, alpha)
}

func evaluateEnergy(tunnel Tunnel, left, middle, right Disk) float64 {
	e1 := evalEnergyTowardsMindisk(tunnel, middle)
	e2 := evalDistanceEnergy(GetDisksDistances(left, middle))
	e3 := evalDistanceEnergy(GetDisksDistances(middle, right))
	return e1 + e2 + e3
}

func evalEnergyTowardsMindisk(tunnel Tunnel, middle Disk) float64 {
	if tunnel.isEnclosingDisk(middle) {
		return 999.
	}
	minDisk := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if middle.radius-minDisk.radius < Delta*0.5 {
		return 0
	} else {
		return math.Pow(middle.radius-minDisk.radius, 1.2)
	}
}

func evalDistanceEnergy(d1, d2 float64) float64 {
	eval := func(d float64) float64 {
		if d < -Eps {
			return math.Pow(math.Abs(1+d), 4)
		} else if d > Delta {
			return math.Pow(math.Abs(d), 4)
		} else {
			return math.Pow(math.Abs(d), 1.2)
		}
	}
	return eval(d1) + eval(d2)
}
