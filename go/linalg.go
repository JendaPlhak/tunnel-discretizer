package main

import (
	"math"
)

var (
	Unit3X = Vec3{1, 0, 0}
	Unit3Y = Vec3{0, 1, 0}
	Unit3Z = Vec3{0, 0, 1}
)

type Vec3 struct {
	x, y, z float64
}

func (v Vec3) IsLinearCombinationOf(u Vec3) bool {
	v.Normalize()
	u.Normalize()
	return SubVec3(v, u).Length() < fError || SubVec3(v.Scaled(-1), u).Length() < fError
}
func (v Vec3) Length() float64 {
	return math.Sqrt(DotVec3(v, v))
}
func (v Vec3) LengthSqr() float64 {
	return DotVec3(v, v)
}
func (v Vec3) Equals(u Vec3) bool {
	return SubVec3(v, u).Length() < fError
}

func (v *Vec3) Normalize() {
	v.Scale(1 / v.Length())
}
func (v Vec3) Normalized() Vec3 {
	v.Normalize()
	return v
}

func AddVec3(u, v Vec3) Vec3 {
	return Vec3{
		u.x + v.x,
		u.y + v.y,
		u.z + v.z,
	}
}
func SubVec3(u, v Vec3) Vec3 {
	return Vec3{
		u.x - v.x,
		u.y - v.y,
		u.z - v.z,
	}
}
func DotVec3(u, v Vec3) float64 {
	return u.x*v.x + u.y*v.y + u.z*v.z
}

// Normal returns a normalized orthogonal vector.
func (v Vec3) Normal() Vec3 {
	n := CrossVec3(v, Unit3Z)
	if n.Length() == 0 {
		return Unit3X
	}
	return n.Normalized()
}
func (v *Vec3) Scale(alpha float64) {
	v.x *= alpha
	v.y *= alpha
	v.z *= alpha
}
func (v Vec3) Scaled(alpha float64) Vec3 {
	v.Scale(alpha)
	return v
}

// Cross returns the cross product of two vectors.
func CrossVec3(a, b Vec3) Vec3 {
	return Vec3{
		a.y*b.z - a.z*b.y,
		a.z*b.x - a.x*b.z,
		a.x*b.y - a.y*b.x,
	}
}

// Cross returns the cross product of two vectors.
func DstVec3(a, b Vec3) float64 {
	return SubVec3(a, b).Length()
}

type Vec2 struct {
	x, y float64
}

func (v Vec2) Length() float64 {
	return math.Sqrt(DotVec2(v, v))
}

func AddVec2(u, v Vec2) Vec2 {
	return Vec2{
		u.x + v.x,
		u.y + v.y,
	}
}
func SubVec2(u, v Vec2) Vec2 {
	return Vec2{
		u.x - v.x,
		u.y - v.y,
	}
}
func DotVec2(u, v Vec2) float64 {
	return u.x*v.x + u.y*v.y
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

// RotAngle3 computes the angle from v1 to v2 with orientation given by the axis.
// The angle is given in interval (0, 2*Pi).
func RotAngle3(axis, v1, v2 Vec3) float64 {
	axis.Normalize()
	alpha := math.Atan2(DotVec3(axis, CrossVec3(v1, v2)), DotVec3(v1, v2))
	if alpha >= 0 {
		return alpha
	}
	return 2*math.Pi + alpha
}

type Mat3x3 struct {
	a00, a01, a02 float64
	a10, a11, a12 float64
	a20, a21, a22 float64
}

func (m Mat3x3) MulVec(u Vec3) Vec3 {
	return Vec3{
		m.a00*u.x + m.a01*u.y + m.a02*u.z,
		m.a10*u.x + m.a11*u.y + m.a12*u.z,
		m.a20*u.x + m.a21*u.y + m.a22*u.z,
	}
}

func (m Mat3x3) Det() float64 {
	//      | a  b  c |
	//  det | d  e  f | = aei + bfg + cdh - ceg - bdi - afh
	//      | g  h  i |
	return m.a00*m.a11*m.a22 + m.a01*m.a12*m.a20 + m.a02*m.a10*m.a21 -
		m.a02*m.a11*m.a20 - m.a01*m.a10*m.a22 - m.a00*m.a12*m.a21
}

func (m Mat3x3) Inversed() Mat3x3 {
	A := Mat3x3{}
	A.a00 = m.a11*m.a22 - m.a21*m.a12
	A.a01 = m.a02*m.a21 - m.a01*m.a22
	A.a02 = m.a01*m.a12 - m.a02*m.a11

	A.a10 = m.a12*m.a20 - m.a10*m.a22
	A.a11 = m.a00*m.a22 - m.a02*m.a20
	A.a12 = m.a10*m.a02 - m.a00*m.a12

	A.a20 = m.a10*m.a21 - m.a20*m.a11
	A.a21 = m.a20*m.a01 - m.a00*m.a21
	A.a22 = m.a00*m.a11 - m.a10*m.a01
	A.Scale(1 / m.Det())
	return A
}
func (m Mat3x3) Transponsed() Mat3x3 {
	return Mat3x3{
		m.a00, m.a10, m.a20,
		m.a01, m.a11, m.a21,
		m.a02, m.a12, m.a22,
	}
}

func (m *Mat3x3) SetRow(i int, u Vec3) {
	switch i {
	case 0:
		m.a00, m.a01, m.a02 = u.x, u.y, u.z
	case 1:
		m.a10, m.a11, m.a12 = u.x, u.y, u.z
	case 2:
		m.a20, m.a21, m.a22 = u.x, u.y, u.z
	}
}
func (m *Mat3x3) Scale(c float64) {
	m.a00, m.a01, m.a02 = c*m.a00, c*m.a01, c*m.a02
	m.a10, m.a11, m.a12 = c*m.a10, c*m.a11, c*m.a12
	m.a20, m.a21, m.a22 = c*m.a20, c*m.a21, c*m.a22
}

func isBasis3D(u, v, w Vec3) bool {
	A := Mat3x3{}
	A.SetRow(0, u)
	A.SetRow(1, v)
	A.SetRow(2, w)
	return math.Abs(A.Det()) > fError
}

// Return the rotation matrix associated with counterclockwise rotation around
// the given axis by theta radians.
func computeRotationMatrix(axis Vec3, theta float64) Mat3x3 {
	a := math.Cos(theta / 2.0)
	v := axis.Normalized().Scaled(-math.Sin(theta / 2.0))
	b, c, d := v.x, v.y, v.z
	aa, bb, cc, dd := a*a, b*b, c*c, d*d
	bc, ad, ac, ab, bd, cd := b*c, a*d, a*c, a*b, b*d, c*d
	return Mat3x3{
		aa + bb - cc - dd, 2 * (bc + ad), 2 * (bd - ac),
		2 * (bc - ad), aa + cc - bb - dd, 2 * (cd + ab),
		2 * (bd + ac), 2 * (cd - ab), aa + dd - bb - cc,
	}
}
