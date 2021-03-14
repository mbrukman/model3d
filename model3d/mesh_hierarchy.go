// Generated from templates/mesh_hierarchy.template

package model3d

// A MeshHierarchy is a tree structure where each node is
// a closed, simple surface, and children are contained
// inside their parents.
//
// Only manifold meshes with no self-intersections can be
// converted into a MeshHierarchy.
type MeshHierarchy struct {
	// Mesh is the root shape of this (sub-)hierarchy.
	Mesh *Mesh

	// MeshSolid is a solid indicating which points are
	// contained in the mesh.
	MeshSolid Solid

	Children []*MeshHierarchy
}

// MeshToHierarchy creates a MeshHierarchy for each
// exterior mesh contained in m.
//
// The mesh m must be manifold and have no
// self-intersections.
func MeshToHierarchy(m *Mesh) []*MeshHierarchy {
	if m.NeedsRepair() {
		panic("mesh needs repair")
	}

	m, inv := misalignMesh(m)
	hierarchy := misalignedMeshToHierarchy(m)
	for i, tree := range hierarchy {
		hierarchy[i] = tree.MapCoords(inv)
	}
	return hierarchy
}

func misalignedMeshToHierarchy(m *Mesh) []*MeshHierarchy {
	pm := newPtrMeshMesh(m)

	var result []*MeshHierarchy

ClosedMeshLoop:
	for pm.Peek() != nil {
		minVertex := pm.Peek()
		pm.IterateCoords(func(c *ptrCoord) {
			if c.X < minVertex.X {
				minVertex = c
			}
		})
		stripped := removeAllConnected(pm, minVertex)
		GroupTriangles(stripped)
		solid := NewColliderSolid(GroupedTrianglesToCollider(stripped))
		strippedMesh := NewMeshTriangles(stripped)
		for _, x := range result {
			if x.MeshSolid.Contains(minVertex.Coord3D) {
				// We know the mesh is a leaf, because if it contained
				// any other mesh, that mesh would have to have a higher
				// minVertex x value, and would not have been added yet.
				x.insertLeaf(strippedMesh, solid, minVertex.Coord3D)
				continue ClosedMeshLoop
			}
		}
		// If we are here, this is a root mesh.
		result = append(result, &MeshHierarchy{
			Mesh:      strippedMesh,
			MeshSolid: solid,
		})
	}

	return result
}

// insertLeaf inserts a mesh into the hierarchy, knowing
// that the mesh is a leaf in the current hierarchy.
func (m *MeshHierarchy) insertLeaf(mesh *Mesh, solid Solid, c Coord3D) {
	v := mesh.VertexSlice()[0]
	for _, child := range m.Children {
		if child.MeshSolid.Contains(v) {
			child.insertLeaf(mesh, solid, c)
			return
		}
	}
	m.Children = append(m.Children, &MeshHierarchy{
		Mesh:      mesh,
		MeshSolid: solid,
	})
}

// FullMesh re-combines the root mesh with all of its
// children.
func (m *MeshHierarchy) FullMesh() *Mesh {
	res := NewMeshTriangles(m.Mesh.TriangleSlice())
	for _, child := range m.Children {
		res.AddMesh(child.FullMesh())
	}
	return res
}

// MapCoords creates a new MeshHierarchy by applying f to
// every coordinate in every mesh.
func (m *MeshHierarchy) MapCoords(f func(Coord3D) Coord3D) *MeshHierarchy {
	res := &MeshHierarchy{
		Mesh: m.Mesh.MapCoords(f),
	}
	res.MeshSolid = NewColliderSolid(MeshToCollider(res.Mesh))
	for _, child := range m.Children {
		res.Children = append(res.Children, child.MapCoords(f))
	}
	return res
}

// Min gets the minimum point of the outer mesh's
// bounding box.
func (m *MeshHierarchy) Min() Coord3D {
	return m.MeshSolid.Min()
}

// Max gets the maximum point of the outer mesh's
// bounding box.
func (m *MeshHierarchy) Max() Coord3D {
	return m.MeshSolid.Max()
}

// Contains checks if c is inside the hierarchy using the
// even-odd rule.
func (m *MeshHierarchy) Contains(c Coord3D) bool {
	if !m.MeshSolid.Contains(c) {
		return false
	}
	for _, child := range m.Children {
		if child.Contains(c) {
			return false
		}
	}
	return true
}

// misalignMesh rotates the mesh by a random angle to
// prevent vertices from directly aligning on the x or
// y axes.
func misalignMesh(m *Mesh) (misaligned *Mesh, inv func(Coord3D) Coord3D) {
	invMapping := map[Coord3D]Coord3D{}
	rot := Rotation(XYZ(0.95177695, 0.26858931, -0.14825794), 0.5037616150469717)
	misaligned = m.MapCoords(func(c Coord3D) Coord3D {
		c1 := rot.Apply(c)
		invMapping[c1] = c
		return c1
	})
	inv = func(c Coord3D) Coord3D {
		if res, ok := invMapping[c]; ok {
			return res
		} else {
			panic("coordinate was not in the misaligned output")
		}
	}
	return misaligned, inv
}

// removeAllConnected strips all segments connected to c
// out of m and returns them as segments.
func removeAllConnected(m *ptrMesh, c *ptrCoord) []*Triangle {
	var result []*Triangle
	resultMap := map[*ptrTriangle]bool{}
	queue := []*ptrTriangle{}
	for _, t := range c.Triangles {
		m.Remove(t)
		resultMap[t] = true
		queue = append(queue, t)
		result = append(result, t.Triangle())
	}
	idx := 0
	for idx < len(queue) {
		t := queue[idx]
		idx++
		for _, c := range t.Coords {
			for _, t1 := range c.Triangles {
				if !resultMap[t1] {
					m.Remove(t1)
					resultMap[t1] = true
					queue = append(queue, t1)
					result = append(result, t1.Triangle())
				}
			}
		}
	}
	return result
}
