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

type Tunnel []Sphere

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

	tunnel := Tunnel{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) > 0 && words[0] == "ATOM" {
			center := NewVec3([]float64{
				parseFloat(words[6]),
				parseFloat(words[7]),
				parseFloat(words[8]),
			})
			radius := parseFloat(words[9])
			fmt.Printf("(%f, %f, %f), %f\n",
				parseFloat(words[6]),
				parseFloat(words[7]),
				parseFloat(words[8]),
				parseFloat(words[9]),
			)
			tunnel = append(tunnel, Sphere{center, radius})
		}
	}
	if err := scanner.Err(); err != nil {
		panicOnError(err)
	}
	return tunnel
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
			Center: [2]float64{cut.center.AtVec(0), cut.center.AtVec(1)},
			Radius: cut.radius,
		})
	}
	minball := minball.ComputeMinball2D(balls)
	minCut := Circle{
		center: NewVec2(minball.Center[:]),
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
	for _, sphere := range t {
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
