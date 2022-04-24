package model3d

import (
	"math"

	"github.com/unixpickle/model3d/numerical"
)

const (
	// ARAPDefaultTolerance is the default convergence
	// tolerance for ARAP.
	// It allows for early convergence stopping.
	ARAPDefaultTolerance = 1e-3

	// ARAPMaxIterations is the default maximum number
	// of iterations for ARAP.
	ARAPMaxIterations = 5000

	// ARAPMinIterations is the default minimum number
	// of iterations before early stopping is allowed.
	ARAPMinIterations = 2
)

type ARAPWeightingScheme int

const (
	// ARAPWeightingCotangent is the default weighting scheme
	// for ARAP from the paper. Unfortunately, it creates a
	// loss function that can potentially become negative.
	ARAPWeightingCotangent ARAPWeightingScheme = iota

	ARAPWeightingAbsCotangent
	ARAPWeightingUniform
)

func (a ARAPWeightingScheme) weight(cot float64) float64 {
	switch a {
	case ARAPWeightingCotangent:
		return cot
	case ARAPWeightingAbsCotangent:
		return math.Abs(cot)
	case ARAPWeightingUniform:
		return 1
	default:
		panic("unknown weighting scheme")
	}
}

// ARAPConstraints maps coordinates from an original mesh
// to destination coordinates on a deformed mesh.
type ARAPConstraints map[Coord3D]Coord3D

// AddAround adds all of the points within r distance of c
// to the constraints, moving them such that c would move
// to target.
func (a ARAPConstraints) AddAround(arap *ARAP, c Coord3D, r float64, target Coord3D) {
	offset := target.Sub(c)
	for _, c1 := range arap.coords {
		if c.Dist(c1) <= r {
			a[c1] = c1.Add(offset)
		}
	}
}

// ARAP implements as-rigid-as-possible deformations for a
// pre-determined mesh.
type ARAP struct {
	coordToIdx map[Coord3D]int
	coords     []Coord3D
	neighbors  [][]int
	weights    [][]float64
	rotWeights [][]float64
	triangles  [][3]int

	tolerance float64
	maxIters  int
	minIters  int
}

// NewARAP creates an ARAP instance for the given mesh
// topology.
//
// The ARAP instance will not hold a reference to m or its
// triangles. Rather, it copies the data as needed.
//
// The instance uses cotangent weights, which are only
// guaranteed to work on meshes with smaller-than-right
// angles.
// For other weighting options, see NewARAPWeighted().
func NewARAP(m *Mesh) *ARAP {
	return NewARAPWeighted(m, ARAPWeightingCotangent, ARAPWeightingCotangent)
}

// NewARAPWeighted creates an ARAP with a specified
// weighting scheme.
//
// The linear weighting scheme is used for linear solves,
// whereas the rotation weighting scheme is used for
// finding rigid transformations.
//
// The ARAP instance will not hold a reference to m or its
// triangles. Rather, it copies the data as needed.
func NewARAPWeighted(m *Mesh, linear, rotation ARAPWeightingScheme) *ARAP {
	coords := m.VertexSlice()
	triangles := m.TriangleSlice()
	a := &ARAP{
		coordToIdx: map[Coord3D]int{},
		coords:     coords,
		neighbors:  make([][]int, len(coords)),
		weights:    make([][]float64, len(coords)),
		rotWeights: make([][]float64, len(coords)),
		triangles:  make([][3]int, 0, len(triangles)),

		tolerance: ARAPDefaultTolerance,
		maxIters:  ARAPMaxIterations,
		minIters:  ARAPMinIterations,
	}

	for i, c := range coords {
		a.coordToIdx[c] = i
	}

	edgeToTri := map[arapEdge][]int{}
	m.Iterate(func(t *Triangle) {
		var tIdxs [3]int
		for i, c := range t {
			tIdxs[i] = a.coordToIdx[c]
		}
		triIdx := len(a.triangles)
		a.triangles = append(a.triangles, tIdxs)

		for i1, c1 := range tIdxs {
			for i2, c2 := range tIdxs {
				if i1 == i2 {
					continue
				}
				if i2 > i1 {
					e := newARAPEdge(c1, c2)
					edgeToTri[e] = append(edgeToTri[e], triIdx)
				}
				var found bool
				for _, n := range a.neighbors[c1] {
					if n == c2 {
						found = true
						break
					}
				}
				if !found {
					a.neighbors[c1] = append(a.neighbors[c1], c2)
				}
			}
		}
	})

	for c1, neighbors := range a.neighbors {
		var weights, rotWeights []float64
		for _, c2 := range neighbors {
			var cotangentSum float64
			for _, t := range edgeToTri[newARAPEdge(c1, c2)] {
				var otherCoord int
				for _, c3 := range a.triangles[t] {
					if c3 != c1 && c3 != c2 {
						otherCoord = c3
						break
					}
				}
				c3Point := a.coords[otherCoord]
				v1 := a.coords[c1].Sub(c3Point)
				v2 := a.coords[c2].Sub(c3Point)
				cosTheta := v1.Normalize().Dot(v2.Normalize())
				cotangentSum += cosTheta / math.Sqrt(math.Max(0, 1-cosTheta*cosTheta))
			}
			weights = append(weights, linear.weight(cotangentSum/2))
			rotWeights = append(rotWeights, rotation.weight(cotangentSum/2))
		}
		a.weights[c1] = weights
		a.rotWeights[c1] = rotWeights
	}

	return a
}

// Tolerance gets the current convergence tolerance.
// Will be ARAPDefaultTolerance by default.
func (a *ARAP) Tolerance() float64 {
	return a.tolerance
}

// SetTolerance changes the convergence tolerance.
// Lower values make the algorithm run longer but arrive
// at more accurate values.
//
// See ARAPDefaultTolerance.
func (a *ARAP) SetTolerance(t float64) {
	a.tolerance = t
}

// MaxIterations gets the maximum allowed number of steps
// before optimization terminates.
func (a *ARAP) MaxIterations() int {
	return a.maxIters
}

// SetMaxIterations sets the maximum allowed number of
// steps before optimization terminates.
func (a *ARAP) SetMaxIterations(m int) {
	a.maxIters = m
}

// MinIterations gets the minimum allowed number of steps
// before optimization terminates.
func (a *ARAP) MinIterations() int {
	return a.minIters
}

// SetMinIterations sets the minimum allowed number of
// steps before optimization terminates.
func (a *ARAP) SetMinIterations(m int) {
	a.minIters = m
}

// Deform creates a new mesh by enforcing constraints on
// some points of the mesh.
func (a *ARAP) Deform(constraints ARAPConstraints) *Mesh {
	outSlice := a.deformMap(newARAPOperator(a, a.indexConstraints(constraints)), nil)
	return a.coordsToMesh(outSlice)
}

// SeqDeformer creates a function that deforms the mesh,
// potentially caching computations across calls.
//
// If coldStart is true, then the previous deformed mesh is
// used as an initial guess for the next deformation. This
// can reduce computation cost during animations.
//
// The returned function is not safe to call from multiple
// Goroutines concurrently.
func (a *ARAP) SeqDeformer(coldStart bool) func(ARAPConstraints) *Mesh {
	var current []Coord3D
	var l *arapOperator
	return func(constraints ARAPConstraints) *Mesh {
		if l == nil {
			l = newARAPOperator(a, a.indexConstraints(constraints))
		} else {
			l.Update(a.indexConstraints(constraints))
		}
		if coldStart {
			current = a.deformMap(l, nil)
		} else {
			current = a.deformMap(l, current)
		}
		return a.coordsToMesh(current)
	}
}

// Laplace deforms the mesh using a simple Laplacian
// heuristic.
//
// This can be used to generate an initial guess for the
// more general Deform() method.
//
// The result maps all old coordinates to new coordinates.
func (a *ARAP) Laplace(constraints ARAPConstraints) map[Coord3D]Coord3D {
	l := newARAPOperator(a, a.indexConstraints(constraints))
	outSlice := a.laplace(l)
	return a.coordsToMap(outSlice)
}

func (a *ARAP) laplace(l *arapOperator) []Coord3D {
	fullL := newARAPOperator(a, nil)
	targets := fullL.Apply(a.coords)
	return l.LinSolve(targets)
}

// DeformMap performs constrained mesh deformation.
//
// The constraints argument maps coordinates from the
// original mesh to their new, fixed locations.
//
// If the initialGuess is specified, it is used for the
// first iteration of the algorithm as a starting point
// for the deformation.
//
// The result maps all old coordinates to new coordinates.
func (a *ARAP) DeformMap(constraints ARAPConstraints,
	initialGuess map[Coord3D]Coord3D) map[Coord3D]Coord3D {
	l := newARAPOperator(a, a.indexConstraints(constraints))
	outSlice := a.deformMap(l, a.initialGuessSlice(initialGuess))
	return a.coordsToMap(outSlice)
}

func (a *ARAP) deformMap(l *arapOperator, initialGuess []Coord3D) []Coord3D {
	if initialGuess == nil {
		initialGuess = a.laplace(l)
	}

	// Enforce constraints on the init.
	currentOutput := l.Unsqueeze(l.Squeeze(initialGuess))

	rotations := a.rotations(currentOutput)
	lastEnergy := a.energy(currentOutput, rotations)
	for iter := 0; iter < a.maxIters; iter++ {
		targets := l.Targets(rotations)
		currentOutput = l.LinSolve(targets)
		rotations = a.rotations(currentOutput)
		energy := a.energy(currentOutput, rotations)
		if iter+1 >= a.minIters && 1-energy/lastEnergy < a.tolerance {
			break
		}
		lastEnergy = energy
	}

	return currentOutput
}

// rotations computes the rotations-of-best-fit for the
// current coordinate positions.
func (a *ARAP) rotations(currentOutput []Coord3D) []Matrix3 {
	rotations := make([]Matrix3, len(a.coords))
	for i, c := range a.coords {
		var covariance Matrix3
		for j, n := range a.neighbors[i] {
			weight := a.rotWeights[i][j]
			origDiff := a.coords[n].Sub(c)
			newDiff := currentOutput[n].Sub(currentOutput[i])
			piece := NewMatrix3Columns(
				origDiff.Scale(newDiff.X),
				origDiff.Scale(newDiff.Y),
				origDiff.Scale(newDiff.Z),
			)
			for i, x := range piece {
				covariance[i] += x * weight
			}
		}
		var u, s, v Matrix3
		covariance.SVD(&u, &s, &v)
		rot := *v.Mul(u.Transpose())
		if rot.Det() < 0 {
			// Scale the column with the smallest singular
			// value.
			idx := 2
			u[idx] *= -1
			u[idx+3] *= -1
			u[idx+6] *= -1
			rot = *v.Mul(u.Transpose())
		}
		rotations[i] = rot
	}
	return rotations
}

// energy computes the ARAP energy to minimize.
func (a *ARAP) energy(currentOutput []Coord3D, rotations []Matrix3) float64 {
	var energy float64
	for i, neighbors := range a.neighbors {
		rotation := rotations[i]
		for j, n := range neighbors {
			w := a.weights[i][j]
			rotated := rotation.MulColumn(a.coords[i].Sub(a.coords[n]))
			diff := currentOutput[i].Sub(currentOutput[n]).Sub(rotated)
			energy += w * diff.Dot(diff)
		}
	}
	return energy
}

// indexConstraints converts the keys to indices.
func (a *ARAP) indexConstraints(constraints ARAPConstraints) map[int]Coord3D {
	res := map[int]Coord3D{}
	for in, out := range constraints {
		if idx, ok := a.coordToIdx[in]; !ok {
			panic("constraint was not in the original mesh")
		} else {
			res[idx] = out
		}
	}
	return res
}

// initialGuessSlice converts a map from old coordinates
// to new ones into a slice of coordinates.
//
// Automatically fills in coordinates that are not
// present.
func (a *ARAP) initialGuessSlice(m map[Coord3D]Coord3D) []Coord3D {
	// Case where default initial guess is used.
	if m == nil {
		return nil
	}

	res := append([]Coord3D{}, a.coords...)
	for k, v := range m {
		if idx, ok := a.coordToIdx[k]; ok {
			res[idx] = v
		} else {
			panic("coordinate used as key was not in the original mesh")
		}
	}
	return res
}

// coordsToMap converts a coordinate slice to a map from
// original mesh coordinates to new ones.
func (a *ARAP) coordsToMap(s []Coord3D) map[Coord3D]Coord3D {
	res := map[Coord3D]Coord3D{}
	for i, c := range s {
		res[a.coords[i]] = c
	}
	return res
}

// coordsToMesh converts a coordinate slice to a mesh.
func (a *ARAP) coordsToMesh(s []Coord3D) *Mesh {
	m := NewMesh()
	for _, t := range a.triangles {
		m.Add(&Triangle{s[t[0]], s[t[1]], s[t[2]]})
	}
	return m
}

// arapOperator implements the Laplace-Beltrami matrix.
//
// By default, it applies the entire matrix.
// However, it also allows for constrained vertices to be
// substituted for their exact values.
type arapOperator struct {
	arap        *ARAP
	constraints map[int]Coord3D

	// Mapping from constrained (reduced) coordinates to
	// full coordinate indices.
	squeezedToFull []int

	// Inverse of squeezedToFull with -1 at constraints.
	fullToSqueezed []int

	chol *numerical.SparseCholesky
}

func newARAPOperator(a *ARAP, constraints map[int]Coord3D) *arapOperator {
	if constraints == nil {
		constraints = map[int]Coord3D{}
	}
	squeezedToFull := make([]int, 0, len(a.coords)-len(constraints))
	fullToSqueezed := make([]int, len(a.coords))
	for i := 0; i < len(a.coords); i++ {
		if _, ok := constraints[i]; !ok {
			fullToSqueezed[i] = len(squeezedToFull)
			squeezedToFull = append(squeezedToFull, i)
		} else {
			fullToSqueezed[i] = -1
		}
	}
	return &arapOperator{
		arap:           a,
		constraints:    constraints,
		squeezedToFull: squeezedToFull,
		fullToSqueezed: fullToSqueezed,
	}
}

// Update updates the constraints.
//
// If the set of constrained vertices remains the same,
// redundant recomputation can be avoided.
func (a *arapOperator) Update(constraints map[int]Coord3D) {
	if len(constraints) != len(a.constraints) {
		*a = *newARAPOperator(a.arap, constraints)
		return
	}
	for k := range constraints {
		if _, ok := a.constraints[k]; !ok {
			*a = *newARAPOperator(a.arap, constraints)
			return
		}
	}
	a.constraints = constraints
}

// LinSolve performs a linear solve for x in Lx=b.
// It is assumed that b and x are unsqueezed (full rank),
// and the constrained rows of b are simply ignored.
func (a *arapOperator) LinSolve(b []Coord3D) []Coord3D {
	if len(a.squeezedToFull) == 0 {
		// All points are constrained.
		return a.Unsqueeze(a.Squeeze(b))
	}

	b = a.Squeeze(b)
	for i, c := range a.SqueezeDelta() {
		b[i] = b[i].Add(c)
	}

	if a.chol == nil {
		a.chol = numerical.NewSparseCholesky(a.squeezedMatrix())
	}

	ins := make([]numerical.Vec3, len(b))
	for i, x := range b {
		ins[i] = x.Array()
	}
	outs := a.chol.ApplyInverseVec3(ins)
	outCoords := make([]Coord3D, len(outs))
	for i, x := range outs {
		outCoords[i] = NewCoord3DArray(x)
	}
	return a.Unsqueeze(outCoords)
}

// Squeeze gets a vector that can be put through the
// operator (i.e. that has constraints removed).
func (a *arapOperator) Squeeze(full []Coord3D) []Coord3D {
	result := make([]Coord3D, len(a.squeezedToFull))
	for i, j := range a.squeezedToFull {
		result[i] = full[j]
	}
	return result
}

// Unsqueeze performs the inverse of squeeze, filling in
// the constrained values as needed.
func (a *arapOperator) Unsqueeze(squeezed []Coord3D) []Coord3D {
	res := make([]Coord3D, len(a.arap.coords))
	for i, s := range a.fullToSqueezed {
		if s != -1 {
			res[i] = squeezed[s]
		} else {
			res[i] = a.constraints[i]
		}
	}
	return res
}

// SqueezeDelta gets the change in the un-constrained
// variables caused by squeezing out the constraints.
//
// This should be added to the other side of linear
// systems to find the correct values.
func (a *arapOperator) SqueezeDelta() []Coord3D {
	res := make([]Coord3D, len(a.squeezedToFull))
	for i, fullIdx := range a.squeezedToFull {
		neighbors := a.arap.neighbors[fullIdx]
		weights := a.arap.weights[fullIdx]
		var result Coord3D
		for j, n := range neighbors {
			w := weights[j]
			if nSqueezed := a.fullToSqueezed[n]; nSqueezed == -1 {
				result = result.Add(a.constraints[n].Scale(w))
			}
		}
		res[i] = result
	}
	return res
}

// Apply applies the Laplace-Beltrami operator to the
// squeezed vector to get another squeezed vector.
func (a *arapOperator) Apply(v []Coord3D) []Coord3D {
	res := make([]Coord3D, len(v))
	for i, fullIdx := range a.squeezedToFull {
		p := v[i]
		neighbors := a.arap.neighbors[fullIdx]
		weights := a.arap.weights[fullIdx]
		var result Coord3D
		for j, n := range neighbors {
			w := weights[j]
			result = result.Add(p.Scale(w))
			if nSqueezed := a.fullToSqueezed[n]; nSqueezed != -1 {
				result = result.Sub(v[nSqueezed].Scale(w))
			}
		}
		res[i] = result
	}
	return res
}

// Targets computes the right-hand side of the Poisson
// equation using rotation matrices.
func (a *arapOperator) Targets(rotations []Matrix3) []Coord3D {
	res := make([]Coord3D, len(a.arap.coords))
	for i, p := range a.arap.coords {
		neighbors := a.arap.neighbors[i]
		weights := a.arap.weights[i]
		var result Coord3D
		for j, n := range neighbors {
			var rotation Matrix3
			m1 := rotations[i]
			m2 := rotations[n]
			for i, x := range m1 {
				rotation[i] = x + m2[i]
			}
			w := weights[j] / 2
			diff := p.Sub(a.arap.coords[n]).Scale(w)
			result = result.Add(rotation.MulColumn(diff))
		}
		res[i] = result
	}
	return res
}

func (a *arapOperator) squeezedMatrix() *numerical.SparseMatrix {
	mat := numerical.NewSparseMatrix(len(a.squeezedToFull))
	for i, fullIdx := range a.squeezedToFull {
		neighbors := a.arap.neighbors[fullIdx]
		weights := a.arap.weights[fullIdx]
		var diagonal float64
		for j, n := range neighbors {
			w := weights[j]
			diagonal += w
			if nSqueezed := a.fullToSqueezed[n]; nSqueezed != -1 {
				mat.Set(i, nSqueezed, -w)
			}
		}
		mat.Set(i, i, diagonal)
	}
	return mat
}

type arapEdge [2]int

func newARAPEdge(i1, i2 int) arapEdge {
	if i1 < i2 {
		return arapEdge{i1, i2}
	} else {
		return arapEdge{i2, i1}
	}
}
