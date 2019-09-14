package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"tunnel-discretizer/go/minball"
)

type Tunnel struct {
	Spheres []Sphere
	Curve   TunnelCurve
}

func MakeTunnel(spheres []Sphere) Tunnel {
	centers := []Vec3{}
	for _, s := range spheres {
		centers = append(centers, s.center)
	}
	return Tunnel{
		Spheres: spheres,
		Curve:   TunnelCurve{centers: centers},
	}
}

func LoadTunnelFromPdbFile(path string) Tunnel {
	file, err := os.Open(path)
	if err != nil {
		panicOnError(err)
	}
	defer file.Close()

	parseFloat := func(str string) float64 {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f
		} else {
			panicOnError(err)
			return 0.
		}
	}

	spheres := []Sphere{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) > 0 && words[0] == "ATOM" {
			center := Vec3{
				parseFloat(words[6]),
				parseFloat(words[7]),
				parseFloat(words[8]),
			}
			radius := parseFloat(words[9])
			fmt.Printf("(%f, %f, %f), %f\n",
				parseFloat(words[6]),
				parseFloat(words[7]),
				parseFloat(words[8]),
				parseFloat(words[9]),
			)
			spheres = append(spheres, Sphere{center, radius})
		}
	}
	if err := scanner.Err(); err != nil {
		panicOnError(err)
	}
	return MakeTunnel(spheres)
}

func (t Tunnel) GetMinimalDisk(point Vec3, normal Vec3) (Disk, bool) {
	diskPlane := MakePlane(point, normal)
	cuts := t.getDiskCuts(diskPlane, point)
	if len(cuts) == 0 {
		return Disk{}, false
	}

	balls := make([]minball.Ball2D, 0, len(cuts))
	for _, cut := range cuts {
		balls = append(balls, minball.Ball2D{
			Center: [2]float64{cut.center.x, cut.center.y},
			Radius: cut.radius,
		})
	}
	minball := minball.ComputeMinball2D(balls)
	minCut := Circle{
		center: Vec2{minball.Center[0], minball.Center[1]},
		radius: minball.Radius,
	}

	return Disk{
		center: diskPlane.transformPointTo3D(minCut.center),
		normal: normal,
		radius: minCut.radius,
	}, true
}

func (t Tunnel) getDiskCuts(diskPlane Plane, point Vec3) []Circle {
	allCuts := []Circle{}
	for _, sphere := range t.Spheres {
		diskCircle, ok := diskPlane.intersectionWithSphere(sphere)
		if ok {
			allCuts = append(allCuts, diskCircle)
		}
	}
	point2D := diskPlane.orthogonalProjectionParametrized(point)
	idx := -1
	for i, c := range allCuts {
		if c.containsPoint(point2D) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return []Circle{}
	}

	cuts := []Circle{}
	toBeAdded := []int{idx}
	added := make([]bool, len(allCuts))

	for len(toBeAdded) != 0 {
		i := toBeAdded[len(toBeAdded)-1]
		toBeAdded = toBeAdded[:len(toBeAdded)-1]
		c := allCuts[i]

		cuts = append(cuts, c)
		added[i] = true

		for j, c2 := range allCuts {
			if !added[j] && c.intersectsWithCircle(c2) {
				toBeAdded = append(toBeAdded, j)
			}
		}
	}
	return cuts
}

func (t Tunnel) isEnclosingDisk(disk Disk) bool {
	diskPlane := MakePlane(disk.center, disk.normal)
	sphere := Sphere{center: disk.center, radius: disk.radius}
	diskCircle, ok := diskPlane.intersectionWithSphere(sphere)
	if !ok {
		panic("The sphere is supposed to intersect with the disk plane")
	}

	cuts := t.getDiskCuts(diskPlane, disk.center)
	if len(cuts) == 0 {
		return false
	}
	for _, circle := range cuts {
		if !diskCircle.containsCircle(circle) {
			return false
		}
	}
	return true
}

func (t Tunnel) getOptimizedDisk(center Vec3, baseNormal Vec3) Disk {
	bestDisk, ok := t.GetMinimalDisk(center, baseNormal)
	if !ok {
		panic("TODO: Better get optimized disk logic!")
	}
	for prevRadius := -1.0; math.Abs(prevRadius-bestDisk.radius) > 0.1; {
		prevRadius = bestDisk.radius
		for i := 0; i < 1000; i++ {
			phi := RandFloat64(0, 2*math.Phi)
			theta := RandFloat64(0, math.Phi/3)
			rotated := bestDisk.getRotatedDisk(theta, phi)
			rotDisk, ok := t.GetMinimalDisk(rotated.center, rotated.normal)
			if ok && rotDisk.radius < bestDisk.radius {
				bestDisk = rotDisk
			}
		}
	}
	return bestDisk
}

func (t Tunnel) getMinimalDisk(pivot Vec3, initialNormal Vec3) Disk {
	bestDisk := t.getOptimizedDisk(pivot, initialNormal)
	initialRadius := bestDisk.radius
	x := 0
	for i := 0; i < 5; i++ {
		theta := (math.Phi / 3) / math.Pow(4, float64(i))

		for foundBetter := true; foundBetter; {
			foundBetter = false
			for phi := 0.; phi < 2*math.Phi; phi += 0.1 {
				x++
				rotDisk := bestDisk.getRotatedDisk(theta, phi)
				disk := t.getOptimizedDisk(rotDisk.center, rotDisk.normal)
				if disk.radius < bestDisk.radius && t.Curve.passesThroughDisk(disk) {
					bestDisk = disk
					sgn := float64(SgnFloat64(DotVec3(bestDisk.normal, initialNormal)))
					bestDisk.normal.Scale(sgn)
					foundBetter = true
				}
			}
		}
	}
	fmt.Println(x)
	fmt.Printf("Initial radius: %f, Optimized: %f\n", initialRadius, bestDisk.radius)
	if DotVec3(bestDisk.normal, initialNormal) <= 0. {
		panic("Minimal disk optimization yilded unexpected disk.")
	}
	return bestDisk
}

type TunnelCurve struct {
	centers []Vec3
}

// Finds whether given `disk` is passed through by a given curve in topological sense.
func (tc TunnelCurve) passesThroughDisk(disk Disk) bool {
	firstPassSgn := 0
	lastPassSgn := 0
	var split *Vec3

	for i := 0; i < len(tc.centers)-1; i++ {
		seg := Segment{tc.centers[i], tc.centers[i+1]}
		if _, ok := seg.intersectionWithDisk(disk); !ok {
			continue
		}
		dir := SubVec3(seg.p2, seg.p1)
		dirSgn := SgnFloat64(DotVec3(disk.normal, dir))
		if disk.containsPoint(tc.centers[i+1]) && split == nil {
			split = &dir
		} else if split != nil {
			sgn := dirSgn * SgnFloat64(DotVec3(disk.normal, *split))
			if sgn > 0 {
				if firstPassSgn == 0 {
					firstPassSgn = dirSgn
				}
				lastPassSgn = dirSgn
			}
			split = nil
		} else {
			if firstPassSgn == 0 {
				firstPassSgn = dirSgn
			}
			lastPassSgn = dirSgn
		}
	}
	return firstPassSgn != 0 && firstPassSgn == lastPassSgn
}
