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
func (v Vec3) RawVector() []float64 {
	return v.VecDense.RawVector().Data
}
func (v Vec3) IsLinearCombinationOf(u Vec3) bool {
	return SubVec3(v.Normalized(), u.Normalized()).Length() < fError
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
	v.VecDense.ScaleVec(1/v.Length(), v.VecDense)
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
func AddVec2(u, v Vec2) Vec2 {
	w := NewVec2(nil)
	w.AddVec(u, v)
	return w
}
func SubVec2(u, v Vec2) Vec2 {
	w := NewVec2(nil)
	w.SubVec(u, v)
	return w
}
func (v Vec2) Length() float64 {
	return math.Sqrt(mat.Dot(v, v))
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

func isBasis3D(u, v, w Vec3) bool {
	A := mat.NewDense(3, 3, nil)
	A.RowView(0) = u.VecDense
}

// Return the rotation matrix associated with counterclockwise rotation around
// the given axis by theta radians.
func computeRotationMatrix(axis Vec3, theta float64) *mat.Dense {
	a := math.Cos(theta / 2.0)
	v := axis.Normalized().Scaled(-math.Sin(theta / 2.0))
	b, c, d := v.AtVec(0), v.AtVec(1), v.AtVec(2)
	aa, bb, cc, dd := a*a, b*b, c*c, d*d
	bc, ad, ac, ab, bd, cd := b*c, a*d, a*c, a*b, b*d, c*d
	return mat.NewDense(3, 3, []float64{
		aa + bb - cc - dd, 2 * (bc + ad), 2 * (bd - ac),
		2 * (bc - ad), aa + cc - bb - dd, 2 * (cd + ab),
		2 * (bd + ac), 2 * (cd - ab), aa + dd - bb - cc,
	})
}
