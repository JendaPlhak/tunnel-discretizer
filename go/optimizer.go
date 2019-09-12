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
	for round := 0; round < 500; round++ {
		fmt.Printf("Beginning round %d, Total energy: %f\n", round, totalEnergy)

		for i := 1; i < len(disks)-1; i++ {
			left, middle, right := disks[i-1], disks[i], disks[i+1]
			disks[i] = optimizeDiskLocally(tunnel, left, middle, right)
		}
		newTotalEnergy := getTotalEnergy(tunnel, disks)
		if totalEnergy-newTotalEnergy < 0.0001 {
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
	movedToMin := middle
	minDisk, ok := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if ok {
		movedToMin = moveTowardsByEps(middle, minDisk)
	}

	alternativePositions := []Disk{
		middle,
		movedToMin,
		moveTowardsByEps(middle, left),
		moveTowardsByEps(middle, right),
		doRandomRotation(tunnel, left, middle, right),
	}
	energies := []DiskWithEnergy{}
	for _, disk := range alternativePositions {
		energies = append(energies, DiskWithEnergy{
			energy: evaluateEnergy(tunnel, left, disk, right),
			disk:   disk,
		})
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
	eps := RandFloat64(0, 3*Eps)
	alpha := math.Min(eps, l) / l
	return disksLinearCombination(source, 1-alpha, target, alpha)
}

func evaluateEnergy(tunnel Tunnel, left, middle, right Disk) float64 {
	e1 := evalEnergyTowardsMindisk(tunnel, middle)
	e2 := evalDistanceEnergy(GetDisksDistances(left, middle))
	e3 := evalDistanceEnergy(GetDisksDistances(middle, right))
	return e1 + e2 + e3
}

func evalEnergyTowardsMindisk(tunnel Tunnel, middle Disk) float64 {
	if !tunnel.isEnclosingDisk(middle) {
		return 999.
	}
	minDisk, ok := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if !ok { // This means that the disk is completelly out-of-proportion or misplaced.
		return 999
	} else if middle.radius-minDisk.radius < Delta*0.5 {
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

func doRandomRotation(tunnel Tunnel, left, middle, right Disk) Disk {
	phi := RandFloat64(0, 2*math.Phi)
	maxTheta := math.Phi / 3
	theta := RandFloat64(0, maxTheta)
	rotated := middle.getRotatedDisk(theta, phi)
	minRot, ok := tunnel.GetMinimalDisk(rotated.center, rotated.normal)
	if ok {
		return minRot
	}
	return middle
}
