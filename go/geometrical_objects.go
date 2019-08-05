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

	d2 := -mat.Dot(lineDir, v) + math.Sqrt(discriminant)
	scaledDir2 := lineDir.Scaled(d2)
	p2 := AddVec3(l.point, scaledDir2)

	return []Vec3{p1, p2}
}

type Line struct {
	point Vec3
	dir   Vec3
}

type Circle struct {
	center Vec2
	radius float64
}

type Plane struct {
	point  Vec3
	normal Vec3

	hasBaseVectors bool
	baseVectors    [2]Vec3

	BaseTransMatrix *mat.Dense
}

func (p Plane) containsPoint(point Vec3) bool {
	return mat.Dot(p.normal, p.point)-mat.Dot(p.normal, point) <= fError
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

func (p Plane) orthogonalProjectionParametrized(point mat.Vector) Vec2 {
	projVec := NewVec3(nil)
	projVec.MulVec(p.getBaseTransMatrix(), point)
	return NewVec2([]float64{
		projVec.AtVec(0), projVec.AtVec(1),
	})
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

func (p Plane) getBaseVectors() (Vec3, Vec3) {
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
