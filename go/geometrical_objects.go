package main

import (
	"math"
)

type Disk struct {
	center Vec3
	normal Vec3
	radius float64
}

func (d Disk) containsPoint(p Vec3) bool {
	v := SubVec3(p, d.center)
	if math.Abs(DotVec3(v, d.normal)) < fError {
		return false
	}
	return v.Length() <= d.radius+fError
}

func (d Disk) isPointInDisksHalfPlane(P Vec3) bool {
	return DotVec3(d.normal, SubVec3(P, d.center)) >= 0
}

func (d Disk) getRotatedDisk(theta, phi float64) Disk {
	axis := computeOrthogonalComplement(d.normal)
	v := computeRotationMatrix(axis[0], theta).MulVec(d.normal)
	v = computeRotationMatrix(axis[1], phi).MulVec(v)
	return Disk{
		center: d.center,
		normal: v.Normalized(),
		radius: d.radius,
	}
}

func (d Disk) intersectionWithLine(line Line) (Vec3, bool) {
	plane := MakePlane(d.center, d.normal)
	P, ok := plane.intersectionWithLine(line)
	if ok && d.containsPoint(P) {
		return P, true
	}
	return Vec3{}, false
}

func (d Disk) hasIntersectionWithLine(line Line) bool {
	_, ok := d.intersectionWithLine(line)
	return ok
}

func disksLinearCombination(d1 Disk, a1 float64, d2 Disk, a2 float64) Disk {
	return Disk{
		center: AddVec3(d1.center.Scaled(a1), d2.center.Scaled(a2)),
		normal: AddVec3(d1.normal.Scaled(a1), d2.normal.Scaled(a2)),
		radius: a1*d1.radius + a2*d2.radius,
	}
}

// GetDisksDistance calculates the segment vertex distances of disks d1 and d2 in projection plane
// given by their normal vectors. The distances can be both positive and negative depending
// on whether the segment vertex of disk d2 is before or after disk d1 with respect to its normal.
func GetDisksDistances(d1, d2 Disk) (float64, float64) {
	var n Vec3
	if d1.normal.IsLinearCombinationOf(d2.normal) {
		n = d1.normal.Normal()
	} else {
		n = CrossVec3(d1.normal, d2.normal)
	}
	// Vector dir1 realizes the line segment from disk's center to the edge of
	// line segment given by a projection of disk d1 to the plane determined by normal n.
	dir1 := CrossVec3(n, d1.normal).Normalized().Scaled(d1.radius)
	dir2 := CrossVec3(n, d2.normal).Normalized().Scaled(d2.radius)

	A1, A2 := AddVec3(d1.center, dir1), AddVec3(d1.center, dir1.Scaled(-1))
	B1, B2 := AddVec3(d2.center, dir2), AddVec3(d2.center, dir2.Scaled(-1))

	// sumDistances := func(A1, A2, B1, B2 Vec3) float64 {
	// 	return SubVec3(A1, B1).Length() + SubVec3(A2, B2).Length()
	// }
	// if sumDistances(A1, A2, B1, B2) > sumDistances(A2, A1, B1, B2) {
	// 	A1, A2 = A2, A1
	// }

	l1, l2 := SubVec3(A1, B1).Length(), SubVec3(A2, B2).Length()
	if !d1.isPointInDisksHalfPlane(B1) {
		l1 = -l1
	}
	if !d1.isPointInDisksHalfPlane(B2) {
		l2 = -l2
	}
	return l1, l2
}

// ProjectPointOntoDisk finds the closest point to the point on the disk's circumference.
func ProjectPointOntoDisk(point Vec3, disk Disk) Vec3 {
	plane := MakePlane(disk.center, disk.normal)
	projPoint := plane.orthogonalProjection(point)
	if diff := SubVec3(projPoint, disk.center); diff.Length() != 0 {
		return AddVec3(diff.Normalized().Scaled(disk.radius), disk.center)
	}
	// If the point is projected exactly on top of the center, chose a point at random.
	v := disk.normal.Normal()
	return AddVec3(v, disk.center)

}

type Sphere struct {
	center Vec3
	radius float64
}

func (s Sphere) containsPoint(p Vec3) bool {
	v := SubVec3(p, s.center)
	return v.Length() <= s.radius
}

// Calculate points of intersection between line and sphere.
// for reference see https://en.wikipedia.org/wiki/Line%E2%80%93sphere_intersection
func (s Sphere) intersectionLine(l Line) []Vec3 {
	lineDir := l.dir.Normalized()
	r := s.radius
	v := SubVec3(l.point, s.center)

	discriminant := math.Pow(DotVec3(lineDir, v), 2) - v.LengthSqr() + math.Pow(r, 2)
	if discriminant < 0.0 {
		return []Vec3{}
	}

	// d is the distance from line.point - line's reference point
	d1 := -DotVec3(lineDir, v) + math.Sqrt(discriminant)
	scaledDir1 := lineDir.Scaled(d1)
	p1 := AddVec3(l.point, scaledDir1)

	d2 := -DotVec3(lineDir, v) - math.Sqrt(discriminant)
	scaledDir2 := lineDir.Scaled(d2)
	p2 := AddVec3(l.point, scaledDir2)

	return []Vec3{p1, p2}
}

type Line struct {
	point Vec3
	dir   Vec3
}

func (l Line) getLinePoint(t float64) Vec3 {
	return AddVec3(l.point, l.dir.Scaled(t))
}

type Segment struct {
	p1, p2 Vec3
}

func (s Segment) intersectionWithDisk(d Disk) (Vec3, bool) {
	l := Line{point: s.p1, dir: SubVec3(s.p2, s.p1)}
	intersection, ok := d.intersectionWithLine(l)
	if ok && s.containsPoint(intersection) {
		return intersection, true
	}
	return Vec3{}, false

}

func (s Segment) containsPoint(p Vec3) bool {
	d1 := SubVec3(s.p1, p).Length()
	d2 := SubVec3(s.p2, p).Length()
	d3 := SubVec3(s.p1, s.p2).Length()
	return math.Abs(d1+d2-d3) < fError
}

type Circle struct {
	center Vec2
	radius float64
}

func (c Circle) containsCircle(other Circle) bool {
	d := SubVec2(c.center, other.center).Length()
	return c.radius+fError >= d+other.radius
}

func (c Circle) containsPoint(point Vec2) bool {
	return SubVec2(c.center, point).Length() <= c.radius
}

func (c Circle) intersectsWithCircle(other Circle) bool {
	return SubVec2(c.center, other.center).Length() <= c.radius+other.radius
}

type Plane struct {
	point  Vec3
	normal Vec3

	hasBaseVectors bool
	baseVectors    [2]Vec3

	BaseTransMatrix *Mat3x3
}

func MakePlane(point, normal Vec3) Plane {
	return Plane{
		point:  point,
		normal: normal.Normalized(),
	}
}

func (p Plane) containsPoint(point Vec3) bool {
	v := SubVec3(point, p.point)
	return math.Abs(DotVec3(p.normal, v)) <= fError
}

func (p Plane) intersectsWithSphere(s Sphere) bool {
	projectedCenter := p.orthogonalProjection(s.center)
	return s.containsPoint(projectedCenter)
}

func (p *Plane) orthogonalProjection(point Vec3) Vec3 {
	// The following code implements following formula:
	// point - dot(point - p.point, p.normal) * p.normal
	pointsDiff := SubVec3(point, p.point)
	dot := DotVec3(pointsDiff, p.normal)
	scaledNormal := p.normal.Scaled(dot)

	projPoint := SubVec3(point, scaledNormal)
	if !p.containsPoint(projPoint) {
		panic("resulting orthonal projection isn't contained in the plane")
	}
	return projPoint
}

func (p *Plane) orthogonalProjectionParametrized(point Vec3) Vec2 {
	v := SubVec3(point, p.point)
	projVec := p.getBaseTransMatrix().MulVec(v)
	return Vec2{
		projVec.x, projVec.y,
	}
}

func (p *Plane) transformPointTo3D(point Vec2) Vec3 {
	v1, v2 := p.getBaseVectors()
	w := AddVec3(v1.Scaled(point.x), v2.Scaled(point.y))
	P := AddVec3(w, p.point)
	if !p.containsPoint(P) {
		panic("transformed point isn't contained in the original plane.")
	}
	return P
}

func (p *Plane) intersectionWithSphere(s Sphere) (Circle, bool) {
	projectedCenter := p.orthogonalProjection(s.center)
	if !s.containsPoint(projectedCenter) {
		return Circle{}, false
	}
	lineDir, _ := p.getBaseVectors()
	line := Line{point: projectedCenter, dir: lineDir}
	intersectionPoints := s.intersectionLine(line)

	v := SubVec3(intersectionPoints[0], line.point)
	r := v.Length()

	return Circle{
		center: p.orthogonalProjectionParametrized(projectedCenter),
		radius: r,
	}, true
}

func (p *Plane) getBaseVectors() (Vec3, Vec3) {
	if !p.hasBaseVectors {
		p.baseVectors = computeOrthogonalComplement(p.normal)
		p.hasBaseVectors = true
	}
	return p.baseVectors[0], p.baseVectors[1]
}

func (p *Plane) getBaseTransMatrix() Mat3x3 {
	if p.BaseTransMatrix == nil {
		v1, v2 := p.getBaseVectors()
		baseMatrix := Mat3x3{}
		baseMatrix.SetRow(0, v1)
		baseMatrix.SetRow(1, v2)
		baseMatrix.SetRow(2, p.normal)
		transMatrix := baseMatrix.Inversed().Transponsed()
		p.BaseTransMatrix = &transMatrix
	}
	return *p.BaseTransMatrix
}

func (p *Plane) intersectionWithLine(line Line) (Vec3, bool) {
	v1, v2 := p.getBaseVectors()
	if !isBasis3D(v1, v2, line.dir) {
		return Vec3{}, false
	}

	A := Mat3x3{}
	A.SetRow(0, v1)
	A.SetRow(1, v2)
	A.SetRow(2, line.dir.Scaled(-1))

	b := SubVec3(line.point, p.point)
	s := A.Inversed().MulVec(b)
	return line.getLinePoint(s.z), true
}
