package model2d

// A Ray is a line originating at a point and extending
// infinitely in some (positive) direction.
type Ray struct {
	Origin    Coord
	Direction Coord
}

// RayCollision is a point where a ray intersects a
// 2-dimensional outline.
type RayCollision struct {
	// The amount of the ray direction to add to the ray
	// origin to hit the point in question.
	//
	// The scale should be non-negative.
	Scale float64

	// The normal pointing outward from the outline at the
	// point of collision.
	Normal Coord

	// Extra contains additional, implementation-specific
	// information about the collision.
	Extra any
}

// A Collider is the outline of a 2-dimensional shape.
// It can count its intersections with a ray, and check if
// any part of the outline is inside a circle.
//
// All methods of a Collider are safe for concurrency.
type Collider interface {
	Bounder

	// RayCollisions enumerates the collisions with a ray.
	// It returns the total number of collisions.
	//
	// f may be nil, in which case this is simply used for
	// counting.
	RayCollisions(r *Ray, f func(RayCollision)) int

	// FirstRayCollision gets the ray collision with the
	// lowest scale.
	//
	// The second return value is false if no collisions
	// were found.
	FirstRayCollision(r *Ray) (collision RayCollision, collides bool)

	// CircleCollision checks if the collider touches a
	// circle with origin c and radius r.
	CircleCollision(c Coord, r float64) bool
}

// ColliderContains checks if a point is within a Collider
// and at least margin away from the border.
//
// If the margin is negative, points are also conatined if
// the point is less than -margin away from the surface.
func ColliderContains(c Collider, coord Coord, margin float64) bool {
	r := &Ray{
		Origin: coord,
		// Random direction; any direction should work, but we
		// want to avoid edge cases and rounding errors.
		Direction: Coord{0.5224892708603626, 0.10494477243214506},
	}
	collisions := c.RayCollisions(r, nil)
	if collisions%2 == 0 {
		if margin < 0 {
			return c.CircleCollision(coord, -margin)
		}
		return false
	}
	return margin <= 0 || !c.CircleCollision(coord, margin)
}

// A SegmentCollider is a 2-dimensional outline which can
// detect if a line segment collides with the outline.
type SegmentCollider interface {
	// SegmentCollision returns true if the segment
	// collides with the outline.
	SegmentCollision(s *Segment) bool
}

// A RectCollider is a 2-dimensional outline which can
// detect if a 2D axis-aligned rectangular area collides
// with the outline.
type RectCollider interface {
	// RectCollision returns true if any part of the
	// outline is inside the rect.
	RectCollision(r *Rect) bool
}

type MultiCollider interface {
	Collider
	SegmentCollider
	RectCollider
}

// MeshToCollider converts a mesh to an efficient
// MultiCollider.
func MeshToCollider(m *Mesh) MultiCollider {
	segs := m.SegmentsSlice()
	GroupSegments(segs)
	return GroupedSegmentsToCollider(segs)
}

// GroupedSegmentsToCollider converts pre-grouped segments
// into an efficient MultiCollider.
// If the segments were not grouped with GroupSegments,
// then the resulting collider may be highly inefficient.
func GroupedSegmentsToCollider(segs []*Segment) MultiCollider {
	if len(segs) == 0 {
		return &joinedMultiCollider{NewJoinedCollider(nil)}
	} else if len(segs) == 1 {
		return segs[0]
	} else {
		mid := len(segs) / 2
		c1 := GroupedSegmentsToCollider(segs[:mid])
		c2 := GroupedSegmentsToCollider(segs[mid:])
		return &joinedMultiCollider{NewJoinedCollider([]Collider{c1, c2})}
	}
}

// BVHToCollider converts a BVH into a MultiCollider in a
// hierarchical way.
func BVHToCollider(b *BVH[*Segment]) MultiCollider {
	if b.Leaf != nil {
		return b.Leaf
	}
	other := make([]Collider, len(b.Branch))
	for i, b1 := range b.Branch {
		other[i] = BVHToCollider(b1)
	}
	return joinedMultiCollider{NewJoinedCollider(other)}
}

////////////////////////////////////////////////////////////
// NOTE: almost all JoinedCollider code was able to be    //
// copied from model3d. This code duplication cannot be   //
// helped, although perhaps `go generate` should be used. //
////////////////////////////////////////////////////////////

// A JoinedCollider wraps multiple other Colliders and
// only passes along rays and circles that enter their
// combined bounding box.
type JoinedCollider struct {
	min       Coord
	max       Coord
	colliders []Collider
}

// NewJoinedCollider creates a JoinedCollider which
// combines zero or more other colliders.
func NewJoinedCollider(other []Collider) *JoinedCollider {
	if len(other) == 0 {
		return &JoinedCollider{}
	}
	res := &JoinedCollider{
		colliders: other,
		min:       other[0].Min(),
		max:       other[0].Max(),
	}
	for _, c := range other[1:] {
		res.min = res.min.Min(c.Min())
		res.max = res.max.Max(c.Max())
	}
	return res
}

func (j *JoinedCollider) Min() Coord {
	return j.min
}

func (j *JoinedCollider) Max() Coord {
	return j.max
}

func (j *JoinedCollider) RayCollisions(r *Ray, f func(RayCollision)) int {
	if !j.rayCollidesWithBounds(r) {
		return 0
	}

	var count int
	for _, c := range j.colliders {
		count += c.RayCollisions(r, f)
	}
	return count
}

func (j *JoinedCollider) FirstRayCollision(r *Ray) (RayCollision, bool) {
	if !j.rayCollidesWithBounds(r) {
		return RayCollision{}, false
	}
	var anyCollides bool
	var closest RayCollision
	for _, c := range j.colliders {
		if collision, collides := c.FirstRayCollision(r); collides {
			if collision.Scale < closest.Scale || !anyCollides {
				closest = collision
				anyCollides = true
			}
		}
	}
	return closest, anyCollides
}

func (j *JoinedCollider) CircleCollision(center Coord, r float64) bool {
	if len(j.colliders) == 0 {
		return false
	}
	if !circleTouchesBounds(center, r, j.min, j.max) {
		return false
	}
	for _, c := range j.colliders {
		if c.CircleCollision(center, r) {
			return true
		}
	}
	return false
}

func (j *JoinedCollider) rayCollidesWithBounds(r *Ray) bool {
	if len(j.colliders) == 0 {
		return false
	}
	minFrac, maxFrac := rayCollisionWithBounds(r, j.min, j.max)
	return minFrac <= maxFrac && maxFrac >= 0
}

type joinedMultiCollider struct {
	*JoinedCollider
}

func (j joinedMultiCollider) SegmentCollision(s *Segment) bool {
	minFrac, maxFrac := rayCollisionWithBounds(&Ray{
		Origin:    s[0],
		Direction: s[1].Sub(s[0]),
	}, j.min, j.max)
	if maxFrac < minFrac || maxFrac < 0 || minFrac > 1 {
		return false
	}
	for _, c := range j.colliders {
		if c.(SegmentCollider).SegmentCollision(s) {
			return true
		}
	}
	return false
}

func (j joinedMultiCollider) RectCollision(r *Rect) bool {
	min := r.MinVal.Max(j.min)
	max := r.MaxVal.Min(j.max)
	if min.Min(max) != min {
		return false
	}
	for _, c := range j.colliders {
		if c.(RectCollider).RectCollision(r) {
			return true
		}
	}
	return false
}
