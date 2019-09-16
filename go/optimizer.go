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

func OptimizeDisks(tunnel Tunnel, disks []Disk) {
	optimizeDisks(tunnel, disks, 1000)
}

func optimizeDisks(tunnel Tunnel, disks []Disk, maxRoundCount int) {
	totalEnergy := getTotalEnergy(tunnel, disks)
	for round := 0; round < maxRoundCount; round++ {
		if round%1 == 0 {
			fmt.Printf("Beginning round %d, Total energy: %f\n", round, totalEnergy)
		}
		for i := 1; i < len(disks)-1; i++ {
			left, middle, right := disks[i-1], disks[i], disks[i+1]
			disks[i] = optimizeDiskLocally(tunnel, left, middle, right)
		}
		newTotalEnergy := getTotalEnergy(tunnel, disks)
		// if totalEnergy-newTotalEnergy == 0 {
		// 	fmt.Printf("Energy stabilized at %f, ending optimization\n", newTotalEnergy)
		// 	break
		// }
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
		movedToMin = moveTowardsByEps(tunnel, middle, minDisk)
	}

	alternativePositions := []Disk{
		middle,
		movedToMin,
		moveTowardsByEps(tunnel, middle, left),
		moveTowardsByEps(tunnel, middle, right),
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

func moveTowardsByEps(tunnel Tunnel, source, target Disk) Disk {
	l := SubVec3(source.normal, target.normal).Length()
	if l == 0. {
		return source
	}
	alpha := RandFloat64(0, 1)
	normal := AddVec3(source.normal.Scaled(alpha), target.normal.Scaled(1-alpha)).Normalized()

	if minDisk, ok := tunnel.GetMinimalDisk(source.center, normal); ok {
		return minDisk
	}
	return source
}

func evaluateEnergy(tunnel Tunnel, left, middle, right Disk) float64 {
	e1 := evalEnergyTowardsMindisk(tunnel, middle)
	e2 := evalDistanceEnergy(GetDisksDistances(left, middle))
	e3 := evalDistanceEnergy(GetDisksDistances(middle, right))
	return e1 + e2 + e3
}

func evalEnergyTowardsMindisk(tunnel Tunnel, middle Disk) float64 {
	return 0.
	if !tunnel.isEnclosingDisk(middle) {
		return 100.
	}
	minDisk, ok := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if !ok { // This means that the disk is completelly out-of-proportion or misplaced.
		return 100
	} else if middle.radius-minDisk.radius < Delta {
		return 0
	} else {
		return math.Pow(middle.radius-minDisk.radius, 1.05)
	}
}

func evalDistanceEnergy(d1, d2 float64) float64 {
	evalEnergy := func(d float64) float64 {
		return math.Exp(1 + math.Abs(d-1))
		if d < 0 {
			return math.Exp(4 + math.Abs(d))
		} else if d > Delta {
			return math.Exp(4 + math.Abs(d))
		} else {
			return math.Pow(math.Abs(d), 2.)
		}
	}
	return math.Max(evalEnergy(d1), evalEnergy(d2))
}

func doRandomRotation(tunnel Tunnel, left, middle, right Disk) Disk {
	phi := RandFloat64(0, 2*math.Phi)
	maxTheta := math.Phi / 2
	theta := RandFloat64(0, maxTheta)
	rotated := middle.getRotatedDisk(theta, phi)
	minRot, ok := tunnel.GetMinimalDisk(rotated.center, rotated.normal)
	if ok && tunnel.Curve.passesThroughDisk(minRot) {
		return minRot
	}
	return middle
}
