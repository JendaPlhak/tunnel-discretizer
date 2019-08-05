package main

import (
	"math"

	"gonum.org/v1/gonum/mat"
)

var (
	Unit3X = NewVec3([]float64{1, 0, 0})
	Unit3Y = NewVec3([]float64{0, 1, 0})
	Unit3Z = NewVec3([]float64{0, 0, 1})
)

type Vec3 struct {
	*mat.VecDense
}

func NewVec3(data []float64) Vec3 {
	return Vec3{mat.NewVecDense(3, data)}
}
func SubVec3(u, v Vec3) Vec3 {
	w := NewVec3(nil)
	w.SubVec(u, v)
	return w
}
func AddVec3(u, v Vec3) Vec3 {
	w := NewVec3(nil)
	w.AddVec(u, v)
	return w
}

// Cross returns the cross product of two vectors.
func CrossVec3(a, b Vec3) Vec3 {
	return NewVec3([]float64{
		a.AtVec(1)*b.AtVec(2) - a.AtVec(2)*b.AtVec(1),
		a.AtVec(2)*b.AtVec(0) - a.AtVec(0)*b.AtVec(2),
		a.AtVec(0)*b.AtVec(1) - a.AtVec(1)*b.AtVec(0),
	})
}

func (v Vec3) Length() float64 {
	return math.Sqrt(mat.Dot(v, v))
}
func (v Vec3) LengthSqr() float64 {
	return mat.Dot(v, v)
}
func (v *Vec3) Normalize() {
	v.ScaleVec(1/v.Length(), v)
}
func (v Vec3) Normalized() Vec3 {
	u := NewVec3(nil)
	u.CloneVec(v)
	u.Normalize()
	return u
}

// Normal returns an orthogonal vector.
func (v Vec3) Normal() Vec3 {
	n := CrossVec3(v, Unit3Z)
	if n.IsZero() {
		return Unit3X
	}
	return n.Normalized()
}

func (v Vec3) Scaled(alpha float64) Vec3 {
	u := NewVec3(nil)
	u.ScaleVec(alpha, v)
	return u
}

type Vec2 struct {
	*mat.VecDense
}

func NewVec2(data []float64) Vec2 {
	return Vec2{mat.NewVecDense(2, data)}
}

func computeOrthogonalComplement(baseVector Vec3) [2]Vec3 {
	u := baseVector.Normal()
	v := CrossVec3(baseVector, u).Normalized()
	return [2]Vec3{u, v}
}

func computeOrthogonalBasis(baseVector Vec3) [3]Vec3 {
	complement := computeOrthogonalComplement(baseVector)
	return [3]Vec3{baseVector.Normalized(), complement[0], complement[1]}
}
