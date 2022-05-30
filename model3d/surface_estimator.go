// Generated from templates/surface_estimator.template

package model3d

import "math"

const (
	DefaultSurfaceEstimatorBisectCount         = 32
	DefaultSurfaceEstimatorNormalSamples       = 40
	DefaultSurfaceEstimatorNormalBisectEpsilon = 1e-4
	DefaultSurfaceEstimatorNormalNoiseEpsilon  = 1e-4
)

// SolidSurfaceEstimator estimates collision points and
// normals on the surface of a solid using search.
type SolidSurfaceEstimator struct {
	// The Solid to estimate the surface of.
	Solid Solid

	// BisectCount, if non-zero, specifies the number of
	// bisections to use in Bisect().
	// Default is DefaultSurfaceEstimatorBisectCount.
	BisectCount int

	// NormalSamples, if non-zero, specifies how many
	// samples to use to approximate normals.
	// Default is DefaultSurfaceEstimatorNormalSamples.
	NormalSamples int

	// RandomSearchNormals can be set to true to disable
	// the binary search to compute normals. Instead, an
	// evolution strategy is performed to estimate the
	// gradient by sampling random points at a distance
	// of NormalNoiseEpsilon.
	RandomSearchNormals bool

	// NormalBisectEpsilon, if non-zero, specifies a small
	// distance to use in a bisection-based method to
	// compute approximate normals.
	//
	// The value must be larger than the distance between
	// the surface and points passed to Normal().
	//
	// Default is DefaultSurfaceEstimatorNormalBisectionEpsilon.
	NormalBisectEpsilon float64

	// NormalNoiseEpsilon, if non-zero, specifies a small
	// distance to use in an evolution strategy when
	// RandomSearchNormals is true.
	//
	// The value must be larger than the distance between
	// the surface and points passed to Normal().
	//
	// Default is DefaultSurfaceEstimatorNormalNoiseEpsilon.
	NormalNoiseEpsilon float64
}

// BisectInterp returns alpha in [min, max] to minimize the
// surface's distance to p1 + alpha * (p2 - p1).
//
// It is assumed that p1 is outside the surface and p2 is
// inside, and that min < max.
func (s *SolidSurfaceEstimator) BisectInterp(p1, p2 Coord3D, min, max float64) float64 {
	d := p2.Sub(p1)
	count := s.bisectCount()
	for i := 0; i < count; i++ {
		f := (min + max) / 2
		if s.Solid.Contains(p1.Add(d.Scale(f))) {
			max = f
		} else {
			min = f
		}
	}
	return (min + max) / 2
}

// Bisect finds the point between p1 and p2 closest to the
// surface, provided that p1 and p2 are on different sides.
func (s *SolidSurfaceEstimator) Bisect(p1, p2 Coord3D) Coord3D {
	var alpha float64
	if s.Solid.Contains(p1) {
		alpha = 1 - s.BisectInterp(p2, p1, 0, 1)
	} else {
		alpha = s.BisectInterp(p1, p2, 0, 1)
	}
	return p1.Add(p2.Sub(p1).Scale(alpha))
}

// Normal computes the normal at a point on the surface.
// The point must be guaranteed to be on the boundary of
// the surface, e.g. from Bisect().
func (s *SolidSurfaceEstimator) Normal(c Coord3D) Coord3D {
	if s.RandomSearchNormals {
		return s.esNormal(c)
	} else {
		return s.bisectNormal(c)
	}
}

func (s *SolidSurfaceEstimator) esNormal(c Coord3D) Coord3D {
	eps := s.normalNoiseEpsilon()
	count := s.normalSamples()
	if count < 1 {
		panic("need at least one sample to estimate normal with random search")
	}

	var normalSum Coord3D
	for i := 0; i < count; i++ {
		delta := NewCoord3DRandUnit()
		c1 := c.Add(delta.Scale(eps))
		if s.Solid.Contains(c1) {
			normalSum = normalSum.Sub(delta)
		} else {
			normalSum = normalSum.Add(delta)
		}
	}
	return normalSum.Normalize()
}

func (s *SolidSurfaceEstimator) bisectNormal(c Coord3D) Coord3D {
	count := s.normalSamples()
	eps := s.normalBisectEpsilon()

	if count < 6 {
		panic("require at least 6 samples to estimate normals with bisection")
	}
	var planeAxes [2]Coord3D
	// Three randomly chosen orthogonal vectors.
	axis1 := XYZ(-0.7107294727984605, -0.12934902142019175, 0.6914712193238857)
	axis2 := XYZ(0.09870891687574183, -0.9915624053549226, -0.08402705526185106)
	axis3 := XYZ(0.696505682837434, 0.008533870423146774, 0.7175005274080017)
	axes := [3]Coord3D{
		axis1.Scale(eps),
		axis2.Scale(eps),
		axis3.Scale(eps),
	}
	var contains [3]bool
	for i, axis := range axes {
		contains[i] = s.Solid.Contains(c.Add(axis))
	}
	for i := 0; i < 2; i++ {
		// Move two vectors towards each other until
		// they are both tangent to the plane.
		v1, c1 := axes[i], contains[i]
		v2, c2 := axes[i+1], contains[i+1]
		if !c1 {
			v1 = v1.Scale(-1)
		}
		if c2 {
			v2 = v2.Scale(-1)
		}
		for j := 0; j < (count-4)/2; j++ {
			mp := v1.Add(v2)
			mp = mp.Scale(eps / mp.Norm())
			if s.Solid.Contains(c.Add(mp)) {
				v1 = mp
			} else {
				v2 = mp
			}
		}
		planeAxes[i] = v1.Add(v2)
		if i == 0 && math.Abs(planeAxes[0].Dot(axes[1])) > math.Abs(planeAxes[0].Dot(axes[0])) {
			// Fix numerical issues when axes[1] is nearly
			// tangent to the surface.
			axes[0], axes[1] = axes[1], axes[0]
			contains[0], contains[1] = contains[1], contains[0]
		}
	}
	res := planeAxes[0].Cross(planeAxes[1]).Normalize()

	if s.Solid.Contains(c.Add(res.Scale(eps))) {
		return res.Scale(-1)
	} else {
		return res
	}
}

func (s *SolidSurfaceEstimator) bisectCount() int {
	if s.BisectCount == 0 {
		return DefaultSurfaceEstimatorBisectCount
	}
	return s.BisectCount
}

func (s *SolidSurfaceEstimator) normalSamples() int {
	if s.NormalSamples == 0 {
		return DefaultSurfaceEstimatorNormalSamples
	}
	return s.NormalSamples
}

func (s *SolidSurfaceEstimator) normalBisectEpsilon() float64 {
	if s.NormalBisectEpsilon == 0 {
		return DefaultSurfaceEstimatorNormalBisectEpsilon
	}
	return s.NormalBisectEpsilon
}

func (s *SolidSurfaceEstimator) normalNoiseEpsilon() float64 {
	if s.NormalNoiseEpsilon == 0 {
		return DefaultSurfaceEstimatorNormalNoiseEpsilon
	}
	return s.NormalNoiseEpsilon
}
