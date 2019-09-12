package main

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

type Disk struct {
	center Vec3
	normal Vec3
	radius float64
}

func (d Disk) Contains(p Vec3) bool {
	v := NewVec3(nil)
	v.SubVec(p, d.center)
	if math.Abs(mat.Dot(v, d.normal)) < fError {
		return false
	}
	return v.Length() <= d.radius+fError
}

func (d Disk) isPointInDisksHalfPlane(P Vec3) bool {
	return mat.Dot(d.normal, SubVec3(P, d.center)) >= 0
}

func (d Disk) getRotatedDisk(theta, phi float64) Disk {
	axis := computeOrthogonalComplement(d.normal)
	v := NewVec3(nil)
	v.MulVec(computeRotationMatrix(axis[0], theta), d.normal)
	v.MulVec(computeRotationMatrix(axis[1], phi), v)
	return Disk{
		center: d.center,
		normal: v.Normalized(),
		radius: d.radius,
	}
}

func (d Disk) hasIntersectionWithLine(line Line) bool {
	plane := MakePlane(d.center, d.normal)
	P, ok := plane.intersectionWithLine(line)
	if !ok {
		return false
	}
	return d.Contains(P)

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

	sumDistances := func(A1, A2, B1, B2 Vec3) float64 {
		return SubVec3(A1, B1).Length() + SubVec3(A2, B2).Length()
	}
	if sumDistances(A1, A2, B1, B2) > sumDistances(A2, A1, B1, B2) {
		A1, A2 = A2, A1
	}

	l1, l2 := SubVec3(A1, B1).Length(), SubVec3(A2, B2).Length()
	if !d1.isPointInDisksHalfPlane(B1) {
		l1 = -l1
	}
	if !d1.isPointInDisksHalfPlane(B2) {
		l2 = -l2
	}
	return l1, l2
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

	discriminant := math.Pow(mat.Dot(lineDir, v), 2) - v.LengthSqr() + math.Pow(r, 2)
	if discriminant < 0.0 {
		return []Vec3{}
	}

	// d is the distance from line.point - line's reference point
	d1 := -mat.Dot(lineDir, v) + math.Sqrt(discriminant)
	scaledDir1 := lineDir.Scaled(d1)
	p1 := AddVec3(l.point, scaledDir1)

	d2 := -mat.Dot(lineDir, v) - math.Sqrt(discriminant)
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

	BaseTransMatrix *mat.Dense
}

func MakePlane(point, normal Vec3) Plane {
	return Plane{
		point:  point,
		normal: normal.Normalized(),
	}
}

func (p Plane) containsPoint(point Vec3) bool {
	v := SubVec3(point, p.point)
	return math.Abs(mat.Dot(p.normal, v)) <= fError
}

func (p Plane) intersectsWithSphere(s Sphere) bool {
	projectedCenter := p.orthogonalProjection(s.center)
	return s.containsPoint(projectedCenter)
}

func (p Plane) orthogonalProjection(point Vec3) Vec3 {
	// The following code implements following formula:
	// point - dot(point - p.point, p.normal) * p.normal
	pointsDiff := SubVec3(point, p.point)
	dot := mat.Dot(pointsDiff, p.normal)
	scaledNormal := p.normal.Scaled(dot)

	projPoint := SubVec3(point, scaledNormal)
	if !p.containsPoint(projPoint) {
		panic("resulting orthonal projection isn't contained in the plane")
	}
	return projPoint
}

func (p *Plane) orthogonalProjectionParametrized(point Vec3) Vec2 {
	v := SubVec3(point, p.point)
	projVec := NewVec3(nil)
	projVec.MulVec(p.getBaseTransMatrix(), v)
	return NewVec2([]float64{
		projVec.AtVec(0), projVec.AtVec(1),
	})
}

func (p Plane) transformPointTo3D(point Vec2) Vec3 {
	v1, v2 := p.getBaseVectors()
	w := NewVec3(nil)
	w.AddVec(v1.Scaled(point.AtVec(0)), v2.Scaled(point.AtVec(1)))
	return AddVec3(w, p.point)
}

func (p Plane) intersectionWithSphere(s Sphere) (Circle, bool) {
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

func (p *Plane) getBaseTransMatrix() *mat.Dense {
	if p.BaseTransMatrix == nil {
		v1, v2 := p.getBaseVectors()
		data := []float64{
			v1.AtVec(0), v1.AtVec(1), v1.AtVec(2),
			v2.AtVec(0), v2.AtVec(1), v2.AtVec(2),
			p.normal.AtVec(0), p.normal.AtVec(1), p.normal.AtVec(2),
		}
		baseMatrix := mat.NewDense(3, 3, data)
		p.BaseTransMatrix = mat.NewDense(3, 3, nil)
		p.BaseTransMatrix.Inverse(baseMatrix)
	}
	return p.BaseTransMatrix
}

func (p *Plane) intersectionWithLine(line Line) (Vec3, bool) {
	v1, v2 := p.getBaseVectors()
	if !isBasis3D(v1, v2, line.dir) {
		return NewVec3(nil), false
	}

	A := mat.NewDense(3, 3, nil)
	A.SetRow(0, v1.RawVector())
	A.SetRow(1, v2.RawVector())
	A.SetRow(2, line.dir.Scaled(-1).RawVector())
	b := SubVec3(line.point, p.point)

	s := NewVec3(nil)
	s.SolveVec(A, b)

	return line.getLinePoint(s.AtVec(2)), true
}
