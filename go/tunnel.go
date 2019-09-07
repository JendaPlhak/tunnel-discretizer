package main

import (
	"bufio"
	"fmt"
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

func (t Tunnel) GetMinimalDisk(point Vec3, normal Vec3) Disk {
	diskPlane := MakePlane(point, normal)
	cuts := t.getDiskCuts(diskPlane, point)

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
	}
}

func (t Tunnel) getIntersectingSpheres(plane Plane, point Vec3) []Sphere {
	containingSpheres := t.getSortedSpheresContainingPoint(point)
	if len(containingSpheres) == 0 {
		return []Sphere{}
	}

	firstIdx := containingSpheres[0].idx
	lastIdx := containingSpheres[len(containingSpheres)-1].idx

	if len(containingSpheres) != lastIdx-firstIdx+1 {
		panic(fmt.Errorf("unexpected number of containing spheres. Expected %d, got %d ",
			lastIdx-firstIdx+1, len(containingSpheres)))
	}

	for i := firstIdx; i >= 0; i-- {
		if !plane.intersectsWithSphere(t[i]) {
			firstIdx = i + 1
			break
		}
	}
	for i := lastIdx; i < len(t); i++ {
		if !plane.intersectsWithSphere(t[i]) {
			lastIdx = i - 1
			break
		}
	}
	return t[firstIdx : lastIdx+1]
}

type sphereWithIndex struct {
	idx    int
	sphere Sphere
}

func (t Tunnel) getSortedSpheresContainingPoint(point Vec3) []sphereWithIndex {
	spheres := []sphereWithIndex{}
	for i, s := range t {
		if s.containsPoint(point) {
			spheres = append(spheres, sphereWithIndex{idx: i, sphere: s})
		}
	}
	return spheres
}

func (t Tunnel) getDiskCuts(diskPlane Plane, point Vec3) []Circle {
	cuts := []Circle{}
	for _, sphere := range t.getIntersectingSpheres(diskPlane, point) {
		cut, ok := diskPlane.intersectionWithSphere(sphere)
		if !ok {
			panic("the plane should be intersecting with the sphere, but isn't")
		}
		cuts = append(cuts, cut)
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
