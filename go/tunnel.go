package main

import (
	"fmt"
)

type Tunnel []Sphere

func (t Tunnel) GetMinimalDisk(point Vec3, normal Vec3) Disk {
	diskPlane := Plane{point: point, normal: normal}
	cuts := []Circle{}

	for _, sphere := range t.getIntersectingSpheres(diskPlane, point) {
		cut, ok := diskPlane.intersectionWithSphere(sphere)
		if !ok {
			panic("the plane should be intersecting with the sphere, but isn't")
		}
		cuts = append(cuts, cut)
	}
	if len(cuts) == 0 {
		panic("disk-tunnel intersection yielded no cuts")
	}
}

func (t Tunnel) getIntersectingSpheres(plane Plane, point Vec3) []Sphere {
	containingSpheres := t.getSortedSpheresContainingPoint(point)
	if len(containingSpheres) == 0 {
		return []Sphere{}
	}

	first_idx := containingSpheres[0].idx
	last_idx := containingSpheres[len(containingSpheres)-1].idx

	if len(containingSpheres) != last_idx-first_idx+1 {
		panic(fmt.Errorf("unexpected number of containing spheres. Expected %d, got %d ",
			last_idx-first_idx+1, len(containingSpheres)))
	}

	for i := first_idx; i >= 0; i-- {
		if !plane.intersectsWithSphere(t[i]) {
			first_idx = i + 1
			break
		}
	}
	for i := last_idx; i < len(t); i++ {
		if !plane.intersectsWithSphere(t[i]) {
			last_idx = i - 1
			break
		}
	}
	return t[first_idx : last_idx+1]
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
