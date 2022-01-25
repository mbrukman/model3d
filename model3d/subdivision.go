package model3d

// LoopSubdivision subdivides the mesh using the Loop
// subdivision rule, creating a smoother surface with
// more triangles.
//
// The mesh is subdivided iters times.
//
// The mesh must not have singular edges.
func LoopSubdivision(m *Mesh, iters int) *Mesh {
	for i := 0; i < iters; i++ {
		m = loopSubdivision(m)
	}
	return m
}

func loopSubdivision(m *Mesh) *Mesh {
	edgePoints := map[Segment]Coord3D{}
	m.Iterate(func(t *Triangle) {
		for _, seg := range t.Segments() {
			if _, ok := edgePoints[seg]; ok {
				continue
			}
			ts := m.Find(seg[0], seg[1])
			if len(ts) != 2 {
				panic("singular edge detected")
			}
			o1 := seg.Other(ts[0])
			o2 := seg.Other(ts[1])
			edgePoints[seg] = seg[0].Add(seg[1]).Scale(3.0 / 8).Add(o1.Add(o2).Scale(1.0 / 8))
		}
	})

	cornerPoints := map[Coord3D]Coord3D{}
	m.getVertexToFace().Range(func(corner Coord3D, tris []*Triangle) bool {
		neighbors := map[Coord3D]bool{}
		for _, t := range tris {
			for _, c := range t {
				if c != corner {
					neighbors[c] = true
				}
			}
		}

		var beta float64
		if len(neighbors) == 3 {
			beta = 3.0 / 16
		} else {
			beta = 3.0 / float64(8*len(neighbors))
		}

		var point Coord3D
		for c := range neighbors {
			point = point.Add(c)
		}
		point = corner.Scale(1 - float64(len(neighbors))*beta).Add(point.Scale(beta))

		cornerPoints[corner] = point
		return true
	})

	res := NewMesh()
	m.Iterate(func(t *Triangle) {
		// Create this triangle:
		//
		//            c1
		//          /    \
		//         m3 -- m1
		//        /  \ /   \
		//       c3-- m2 --c2
		//
		c1, c2, c3 := cornerPoints[t[0]], cornerPoints[t[1]], cornerPoints[t[2]]
		m1 := edgePoints[NewSegment(t[0], t[1])]
		m2 := edgePoints[NewSegment(t[1], t[2])]
		m3 := edgePoints[NewSegment(t[2], t[0])]

		res.Add(&Triangle{m1, m2, m3})
		res.Add(&Triangle{c1, m1, m3})
		res.Add(&Triangle{m1, c2, m2})
		res.Add(&Triangle{m3, m2, c3})
	})
	return res
}

// SubdivideEdges sub-divides each edge into n sub-edges
// and fills each original triangle with n^2 coplanar
// triangles.
func SubdivideEdges(m *Mesh, n int) *Mesh {
	if n < 1 {
		panic("number of sub-edges must be at least 1")
	}

	side1 := make([]Coord3D, n+1)
	side2 := make([]Coord3D, n+1)
	wideLine := make([]Coord3D, n+1)
	narrowLine := make([]Coord3D, n)

	res := NewMesh()
	m.Iterate(func(t *Triangle) {
		divideSegment(t[0], t[1], side1)
		divideSegment(t[0], t[2], side2)

		for i := 0; i < n; i++ {
			wideLine = wideLine[:i+2]
			narrowLine = narrowLine[:i+1]
			divideSegment(side1[i], side2[i], narrowLine)
			divideSegment(side1[i+1], side2[i+1], wideLine)
			for i := 0; i < len(narrowLine); i++ {
				res.Add(&Triangle{narrowLine[i], wideLine[i], wideLine[i+1]})
				if i > 0 {
					res.Add(&Triangle{narrowLine[i], narrowLine[i-1], wideLine[i]})
				}
			}
		}
	})
	return res
}

func divideSegment(c1, c2 Coord3D, result []Coord3D) {
	if len(result) == 1 {
		result[0] = c1
		return
	}
	// Rounding errors may make the result depend on the order
	// of c1 and c2, so we explicitly order them.
	seg := NewSegment(c1, c2)
	if seg[0] != c1 {
		divideSegment(c2, c1, result)
		for i := 0; i < len(result)/2; i++ {
			result[i], result[len(result)-(i+1)] = result[len(result)-(i+1)], result[i]
		}
	} else {
		result[0] = c1
		result[len(result)-1] = c2
		for i := 1; i+1 < len(result); i++ {
			t := float64(i) / float64(len(result)-1)
			result[i] = c1.Scale(1 - t).Add(c2.Scale(t))
		}
	}
}

// A Subdivider is used for sub-dividing triangles in a
// mesh to add levels of detail where it is needed in a
// model.
//
// To achieve this, it tracks line segments that are to be
// split, and then replaces triangles in the mesh such
// that the given lines are split and the midpoint is
// replaced with a more accurate value.
type Subdivider struct {
	lines map[Segment]bool
}

// NewSubdivider creates an empty Subdivider.
func NewSubdivider() *Subdivider {
	return &Subdivider{lines: map[Segment]bool{}}
}

// Add adds a line segment that needs to be split.
func (s *Subdivider) Add(p1, p2 Coord3D) {
	s.lines[NewSegment(p1, p2)] = true
}

// AddFiltered adds line segments for which f returns
// true.
func (s *Subdivider) AddFiltered(m *Mesh, f func(p1, p2 Coord3D) bool) {
	visited := map[Segment]bool{}
	m.Iterate(func(t *Triangle) {
		segs := [3]Segment{
			NewSegment(t[0], t[1]),
			NewSegment(t[1], t[2]),
			NewSegment(t[2], t[0]),
		}
		for _, seg := range segs {
			if !visited[seg] {
				visited[seg] = true
				if f(seg[0], seg[1]) {
					s.Add(seg[0], seg[1])
				}
			}
		}
	})
}

// NumSegments returns the current number of segments
// added to s.
func (s *Subdivider) NumSegments() int {
	return len(s.lines)
}

// Subdivide modifies the mesh by replacing triangles
// whose sides are affected by subdivision.
//
// The midpointFunc computes a 3D coordinate that should
// replace the midpoint of a given line segment from the
// original mesh.
func (s *Subdivider) Subdivide(mesh *Mesh, midpointFunc func(p1, p2 Coord3D) Coord3D) {
	midpoints := map[Segment]Coord3D{}
	for segment := range s.lines {
		midpoints[segment] = midpointFunc(segment[0], segment[1])
	}
	mesh.Iterate(func(t *Triangle) {
		segs := [3]Segment{
			NewSegment(t[0], t[1]),
			NewSegment(t[1], t[2]),
			NewSegment(t[2], t[0]),
		}
		var splits [3]bool
		var numSplits int
		for i, seg := range segs {
			splits[i] = s.lines[seg]
			if splits[i] {
				numSplits++
			}
		}
		if numSplits == 1 {
			for i, seg := range segs {
				if splits[i] {
					subdivideSingle(mesh, t, seg, midpoints[seg])
					break
				}
			}
		} else if numSplits == 2 {
			if !splits[0] {
				subdivideDouble(mesh, t, segs[1], segs[2], midpoints)
			} else if !splits[1] {
				subdivideDouble(mesh, t, segs[0], segs[2], midpoints)
			} else {
				subdivideDouble(mesh, t, segs[0], segs[1], midpoints)
			}
		} else if numSplits == 3 {
			subdivideTriple(mesh, t, midpoints)
		}
	})
}

func subdivideSingle(mesh *Mesh, t *Triangle, splitSeg Segment, midpoint Coord3D) {
	p3 := t[0]
	if p3 == splitSeg[0] || p3 == splitSeg[1] {
		p3 = t[1]
		if p3 == splitSeg[0] || p3 == splitSeg[1] {
			p3 = t[2]
		}
	}
	replaceTriangle(mesh, t,
		&Triangle{splitSeg[0], splitSeg.Mid(), p3}, &Triangle{splitSeg[0], midpoint, p3},
		&Triangle{p3, splitSeg.Mid(), splitSeg[1]}, &Triangle{p3, midpoint, splitSeg[1]})
}

func subdivideDouble(mesh *Mesh, t *Triangle, seg1, seg2 Segment,
	midpoints map[Segment]Coord3D) {
	mp1 := midpoints[seg1]
	mp2 := midpoints[seg2]
	shared := seg1.union(seg2)
	unshared1, unshared2 := seg1.inverseUnion(seg2)
	replaceTriangle(mesh, t,
		&Triangle{shared, seg1.Mid(), seg2.Mid()}, &Triangle{shared, mp1, mp2},
		&Triangle{seg1.Mid(), seg2.Mid(), unshared1}, &Triangle{mp1, mp2, unshared1},
		&Triangle{unshared1, unshared2, seg2.Mid()}, &Triangle{unshared1, unshared2, mp2})
}

func subdivideTriple(mesh *Mesh, t *Triangle, midpoints map[Segment]Coord3D) {
	seg1 := NewSegment(t[0], t[1])
	seg2 := NewSegment(t[1], t[2])
	seg3 := NewSegment(t[2], t[0])
	mp1 := midpoints[seg1]
	mp2 := midpoints[seg2]
	mp3 := midpoints[seg3]
	replaceTriangle(mesh, t,
		&Triangle{seg1.Mid(), t[1], seg3.Mid()}, &Triangle{mp1, t[1], mp2},
		&Triangle{seg2.Mid(), t[2], seg3.Mid()}, &Triangle{mp2, t[2], mp3},
		&Triangle{seg3.Mid(), t[0], seg1.Mid()}, &Triangle{mp3, t[0], mp1},
		&Triangle{seg1.Mid(), seg2.Mid(), seg3.Mid()}, &Triangle{mp1, mp2, mp3})
}

func replaceTriangle(mesh *Mesh, original *Triangle, ts ...*Triangle) {
	if len(ts)%2 != 0 {
		panic("must pass each sub-divided triangle followed by the new triangle")
	}
	mesh.Remove(original)

	norm := original.Normal()
	for i := 0; i < len(ts); i += 2 {
		t := ts[i+1]
		// Make sure the triangle is facing the same way.
		if ts[i].Normal().Dot(norm) < 0 {
			t[1], t[2] = t[2], t[1]
		}
		mesh.Add(t)
	}
}
