// Generated from templates/metaball_test.template

package model3d

import (
	"math"
	"testing"
)

func TestMetaballEquivalence(t *testing.T) {
	testEquiv := func(t *testing.T, m Metaball, radius float64, s SDF) {
		solid := MetaballSolid(nil, radius, m)

		mesh := MarchingCubesSearch(solid, 0.02, 8)

		realSDF := MeshToSDF(mesh)

		min := realSDF.Min().Min(s.Min())
		max := realSDF.Max().Max(s.Max())

		for i := 0; i < 1000; i++ {
			c := NewCoord3DRandBounds(min, max)
			actual := realSDF.SDF(c)
			expected := s.SDF(c)
			if math.Abs(actual-expected) > 0.05 {
				t.Errorf("point %v: expected SDF %f but got SDF %f", c, expected, actual)
			}
		}
	}

	t.Run("Sphere", func(t *testing.T) {
		c := XY(1.7, -0.3)
		testEquiv(
			t,
			&Sphere{Center: c, Radius: 1.0},
			0.5,
			&Sphere{Center: c, Radius: 1.5},
		)
	})
}

func TestMetaballEquivalenceVecScale(t *testing.T) {
	obj := &Sphere{
		Center: XY(1, 2),
		Radius: 1.0,
	}
	smaller := *obj
	smaller.Radius = 0.3 // mix this radius with metaball radius
	scale := XYZ(0.25, 0.5, 1.0)

	scaledSolid := VecScaleSolid(obj, scale)
	scaledMB := VecScaleMetaball(&smaller, scale)
	scaledMBSolid := MetaballSolid(nil, 0.7, scaledMB)

	expectedMesh := MarchingCubesSearch(scaledSolid, 0.02, 8)
	actualMesh := MarchingCubesSearch(scaledMBSolid, 0.02, 8)

	expectedSDF := MeshToSDF(expectedMesh)
	actualSDF := MeshToSDF(actualMesh)

	min := actualSDF.Min().Min(expectedSDF.Min())
	max := actualSDF.Max().Max(expectedSDF.Max())

	for i := 0; i < 1000; i++ {
		c := NewCoord3DRandBounds(min, max)
		actual := actualSDF.SDF(c)
		expected := expectedSDF.SDF(c)
		if math.Abs(actual-expected) > 0.05 {
			t.Errorf("point %v: expected SDF %f but got SDF %f", c, expected, actual)
		}
	}
}

func TestMetaballOverlaid(t *testing.T) {
	// If we outset by r=1, then 1/r^4 = 1.0.
	// Then if we have two metaballs, we want r s.t.
	// 1/r^4 = 2.0, r=2^(1/4)=1.18920711500272106671.
	const expectedOutset = 1.18920711500272106671

	sphere := &Sphere{Center: XY(1, 2), Radius: 1.73}
	mb := MetaballSolid(QuarticMetaballFalloffFunc, 1.0, sphere, sphere)
	expMin := sphere.Min().AddScalar(-expectedOutset)
	expMax := sphere.Max().AddScalar(expectedOutset)
	if mb.Min().Dist(expMin) > 1e-4 ||
		mb.Max().Dist(expMax) > 1e-4 {
		t.Fatalf("expected bounds %v, %v but got %v, %v", expMin, expMax, mb.Min(), mb.Max())
	}
	if mb.Min().Min(expMin) != mb.Min() ||
		mb.Max().Max(expMax) != mb.Max() {
		t.Error("approximate bound is not conservative")
	}
}
