package main

import (
	"fmt"
	"math"
)

type OrientedDisk struct {
	Disk
	up Vec3
}

func (d OrientedDisk) getUp() Vec3 {
	return d.up
}
func (d OrientedDisk) getDown() Vec3 {
	return d.up.Scaled(-1)
}
func (d OrientedDisk) getRight() Vec3 {
	return CrossVec3(d.normal, d.up)
}
func (d OrientedDisk) getLeft() Vec3 {
	return CrossVec3(d.normal, d.getUp()).Scaled(-1)
}

func (d OrientedDisk) getPivotUp() Vec3 {
	return AddVec3(d.up.Scaled(d.radius), d.center)
}

func makeOrientedDisk(disk Disk, up Vec3) OrientedDisk {
	return OrientedDisk{
		Disk: disk,
		up:   up.Normalized(),
	}
}

func OptimizeDisks(tunnel Tunnel, initialDisks []Disk) {
	if len(initialDisks) == 0 {
		return
	}
	firstDisk := initialDisks[0]
	firstPivot := ProjectPointOntoDisk(
		AddVec3(Vec3{0, firstDisk.radius, 0}, firstDisk.center),
		firstDisk,
	)
	disks := []OrientedDisk{
		makeOrientedDisk(firstDisk, SubVec3(firstPivot, firstDisk.center)),
	}
	for i, disk := range initialDisks {
		disks = append(disks, makeOrientedDisk(
			disk,
			SubVec3(
				ProjectPointOntoDisk(disks[i-1].getPivotUp(), disk),
				disk.center,
			),
		))
	}
	optimizeDisks(tunnel, disks, 10000)
}

func optimizeDisks(tunnel Tunnel, disks []OrientedDisk, maxRoundCount int) {
	totalEnergy := getTotalEnergy(tunnel, disks)
	for round := 0; round < maxRoundCount; round++ {
		if round%1 == 0 {
			fmt.Printf("Beginning round %d, Total energy: %f\n", round, totalEnergy)
		}
		for i := 1; i < len(disks)-1; i++ {
			// left, middle, right := disks[i-1], disks[i], disks[i+1]
			// disks[i] = optimizeDiskLocally(tunnel, left, middle, right)
		}
		newTotalEnergy := getTotalEnergy(tunnel, disks)
		// if totalEnergy-newTotalEnergy == 0 {
		// 	fmt.Printf("Energy stabilized at %f, ending optimization\n", newTotalEnergy)
		// 	break
		// }
		totalEnergy = newTotalEnergy
	}
}

func getTotalEnergy(tunnel Tunnel, disks []OrientedDisk) float64 {
	totalEnergy := 0.
	for i := 1; i < len(disks)-1; i++ {
		totalEnergy += evaluateEnergy(tunnel, disks, i)
	}
	return totalEnergy
}

// func optimizeDiskLocally(tunnel Tunnel, left, middle, right Disk) Disk {
// 	movedToMin := middle
// 	minDisk, ok := tunnel.GetMinimalDisk(middle.center, middle.normal)
// 	if ok {
// 		movedToMin = moveTowardsByEps(tunnel, middle, minDisk)
// 	}

// 	alternativePositions := []Disk{
// 		middle,
// 		movedToMin,
// 		moveTowardsByEps(tunnel, middle, left),
// 		moveTowardsByEps(tunnel, middle, right),
// 		doRandomRotation(tunnel, left, middle, right),
// 	}
// 	energies := []DiskWithEnergy{}
// 	for _, disk := range alternativePositions {
// 		energies = append(energies, DiskWithEnergy{
// 			energy: evaluateEnergy(tunnel, left, disk, right),
// 			disk:   disk,
// 		})
// 	}
// 	return getMinEnergyDisk(energies)
// }

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

func evaluateEnergy(tunnel Tunnel, disks []OrientedDisk, i int) float64 {
	left, middle, right := disks[i-1].Disk, disks[i].Disk, disks[i+1].Disk
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

// func evalDistanceEnergy(d1, d2 float64) float64 {
// 	evalEnergy := func(d float64) float64 {
// 		if d < 0 {
// 			return 100*math.Abs(d) + 50
// 		} else if d < 0.1*Delta {
// 			k := -45 / 0.1 * Delta
// 			return k*d + 50
// 		} else if d < 0.8*Delta {
// 			return 0
// 		} else if d < 0.9*Delta {
// 			return 0
// 		} else if d < Delta {
// 			return (450/Delta)*d - 400
// 		} else {
// 			return (100/Delta)*d - 50
// 		}
// 	}
// 	return (evalEnergy(d1) + evalEnergy(d2)) / 2
// }

func evalDistanceEnergy(d1, d2 float64) float64 {
	evalEnergy := func(d float64) float64 {
		if d < 0.1*Delta {
			return math.Exp(4 + math.Abs(d))
		} else if d > Delta {
			return math.Exp(4 + math.Abs(d))
		} else {
			return math.Pow(math.Abs(d-0.5*Delta), 2.)
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
