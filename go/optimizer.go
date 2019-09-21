package main

import (
	"fmt"
	"math"
)

type OrientedDisk struct {
	Disk
	up Vec3
}

func makeOrientedDisk(disk Disk, up Vec3) OrientedDisk {
	return OrientedDisk{
		Disk: disk,
		up:   up.Normalized(),
	}
}

func (d OrientedDisk) getUp() Vec3 {
	return d.up
}
func (d OrientedDisk) getDown() Vec3 {
	return d.up.Scaled(-1)
}
func (d OrientedDisk) getRight() Vec3 {
	return CrossVec3(d.normal, d.getUp())
}
func (d OrientedDisk) getLeft() Vec3 {
	return CrossVec3(d.normal, d.getUp()).Scaled(-1)
}

func (d OrientedDisk) getDirForIdx(idx dirIdx) Vec3 {
	switch idx {
	case 0:
		return d.getUp()
	case 1:
		return d.getRight()
	case 2:
		return d.getDown()
	case 3:
		return d.getLeft()
	default:
		panic("invalid dir index")
	}
}

func (d OrientedDisk) getPivot(idx dirIdx) Vec3 {
	return AddVec3(d.getDirForIdx(idx).Scaled(d.radius), d.center)
}

func (d OrientedDisk) getPivotUp() Vec3 {
	return AddVec3(d.up.Scaled(d.radius), d.center)
}

type dirIdx int

const (
	UpDirIdx    dirIdx = 0
	RightDirIdx dirIdx = 1
	DownDirIdx  dirIdx = 2
	LeftDirIdx  dirIdx = 3
)

func oppositeDir(idx dirIdx) dirIdx {
	return (idx + 2) % 4
}

func followingDir(idx dirIdx) dirIdx {
	return (idx + 1) % 4
}

type ForceField struct {
	Top, Bottom, Right, Left Vec3
}

func (ff *ForceField) getForcePtr(idx dirIdx) *Vec3 {
	switch idx {
	case 0:
		return &ff.Top
	case 1:
		return &ff.Right
	case 2:
		return &ff.Bottom
	case 3:
		return &ff.Left
	default:
		panic("invalid dir index")
	}
}

func (ff ForceField) getForce(idx dirIdx) Vec3 {
	return *ff.getForcePtr(idx)
}

func (ff *ForceField) setForce(idx dirIdx, force Vec3) {
	*ff.getForcePtr(idx) = force
}

func (ff *ForceField) addForceField(other ForceField) {
	for _, idx := range []dirIdx{UpDirIdx, RightDirIdx, DownDirIdx, LeftDirIdx} {
		*ff.getForcePtr(idx) = AddVec3(ff.getForce(idx), other.getForce(idx))
	}
}

// getPrecedingDir returns an index of one of the four major directions,
// specifically it will be the one that has the quadrant that contains the point
// directly to its right.
func (d OrientedDisk) getPrecedingDir(point Vec3) dirIdx {
	if !d.Disk.containsPoint(point) {
		panic("disk doesn't contain the point its quadrant we want to find")
	}
	v := SubVec3(point, d.Disk.center)
	alpha := RotAngle3(d.normal, d.getUp(), v)
	quad := math.Floor(alpha / (math.Pi / 2))
	return dirIdx(int(quad))
}

func (d OrientedDisk) getForceField(point, force Vec3) ForceField {
	idx1 := d.getPrecedingDir(point)
	idx2 := followingDir(idx1)

	P, Q := d.getPivot(idx1), d.getPivot(idx2)
	s := Segment{P, Q}
	X, ok := s.orthoProjPoint(point)
	if !ok {
		panic("Failed to project the force point")
	}
	dstPQ := DstVec3(P, Q)

	v := SubVec3(point, d.Disk.center)
	beta := RotAngle3(CrossVec3(d.getUp(), d.normal), d.getUp(), v)
	F := force.Length()

	f1 := d.normal.Scaled(F * (1 - DstVec3(P, X)/dstPQ) * math.Cos(beta))
	f2 := d.normal.Scaled(F * (1 - DstVec3(Q, X)/dstPQ) * math.Cos(beta))

	forceField := ForceField{}
	forceField.setForce(idx1, f1)
	forceField.setForce(idx2, f2)
	forceField.setForce(oppositeDir(idx1), f1.Scaled(-1))
	forceField.setForce(oppositeDir(idx2), f2.Scaled(-2))
	return forceField
}

func getForceField(d1, d2 OrientedDisk) ForceField {
	dst1, dst2 := GetDisksDistances(d1.Disk, d2.Disk)
	s1, s2 := GetDisksSegments(d1.Disk, d2.Disk)
	v1 := SubVec3(s2.p1, s1.p1)
	v2 := SubVec3(s2.p2, s1.p2)

	ff := d1.getForceField(s1.p1, getForceVector(v1, dst1))
	ff.addForceField(d1.getForceField(s1.p2, getForceVector(v2, dst2)))
	return ff
}

func getForceVector(v Vec3, dst float64) Vec3 {
	energy := float64(SgnFloat64(dst)) * evalDistanceEnergy(dst)
	return v.Normalized().Scaled(energy)
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
				ProjectPointOntoDisk(disks[i-2].getPivotUp(), disk),
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
		totalEnergy += evaluateEnergy2(tunnel, disks, i)
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
	e2 := evalDistanceEnergy2(GetDisksDistances(left, middle))
	e3 := evalDistanceEnergy2(GetDisksDistances(middle, right))
	return e1 + e2 + e3
}

func evaluateEnergy2(tunnel Tunnel, disks []OrientedDisk, i int) float64 {
	// left, middle, right := disks[i-1], disks[i], disks[i+1]
	// ff1 := getForceField(left, middle)
	// ff1 := getForceField(left, right)
	return 0
}

func evalEnergyTowardsMindisk(tunnel Tunnel, middle Disk) float64 {
	return 0.
	if !tunnel.isEnclosingDisk(middle) {
		return 100.
	}
	minDisk, ok := tunnel.GetMinimalDisk(middle.center, middle.normal)
	if !ok { // This means that the disk is completely out-of-proportion or misplaced.
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

func evalDistanceEnergy(d float64) float64 {
	if d < 0.1*Delta {
		return math.Exp(4 + math.Abs(d))
	} else if d > Delta {
		return math.Exp(4 + math.Abs(d))
	} else {
		return math.Pow(math.Abs(d-0.5*Delta), 2.)
	}
}

func evalDistanceEnergy2(d1, d2 float64) float64 {
	return math.Max(evalDistanceEnergy(d1), evalDistanceEnergy(d2))
}

func doRandomRotation(tunnel Tunnel, left, middle, right Disk) Disk {
	phi := RandFloat64(0, 2*math.Pi)
	maxTheta := math.Pi / 2
	theta := RandFloat64(0, maxTheta)
	rotated := middle.getRotatedDisk(theta, phi)
	minRot, ok := tunnel.GetMinimalDisk(rotated.center, rotated.normal)
	if ok && tunnel.Curve.passesThroughDisk(minRot) {
		return minRot
	}
	return middle
}
