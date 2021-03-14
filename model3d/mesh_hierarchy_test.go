// Generated from templates/mesh_hierarchy_test.template

package model3d

import (
	"compress/gzip"
	"os"
	"testing"
)

func TestMeshHierarchy(t *testing.T) {
	mesh, numHier, knownDepth := hierarchyTestingMesh(t)
	hierarchy := MeshToHierarchy(mesh)

	// Specific tests for this hierarchy.
	if len(hierarchy) != numHier {
		t.Errorf("expected %d separate roots but found %d", numHier, len(hierarchy))
	}
	if depth := measureHierarchyDepth(hierarchy); depth != knownDepth {
		t.Errorf("expected %d nested meshes but found %d", knownDepth, depth)
	}

	// Make sure all vertices are preserved.
	flatCount := len(mesh.VertexSlice())
	hierarchyCount := countHierarchyVertices(hierarchy)
	if flatCount != hierarchyCount {
		t.Errorf("expected %v vertices but got %v", flatCount, hierarchyCount)
	}

	// Make sure child containment is preserved.
	for _, h := range hierarchy {
		validateHierarchyContainment(t, h)
	}
	rawSolid := NewColliderSolid(MeshToCollider(mesh))
	for i := 0; i < 10000; i++ {
		c := NewCoord3DRandBounds(rawSolid.Min(), rawSolid.Max())
		contained := rawSolid.Contains(c)
		hContained := false
		for _, h := range hierarchy {
			if h.Contains(c) {
				hContained = true
			}
		}
		if contained != hContained {
			t.Errorf("point %v should have contained=%v but have %v", c, contained, hContained)
		}
	}
}

func countHierarchyVertices(hierarchies []*MeshHierarchy) int {
	var res int
	for _, child := range hierarchies {
		res += len(child.Mesh.VertexSlice())
		res += countHierarchyVertices(child.Children)
	}
	return res
}

func measureHierarchyDepth(hierarchies []*MeshHierarchy) int {
	var result int
	for _, h := range hierarchies {
		depth := measureHierarchyDepth(h.Children) + 1
		if depth > result {
			result = depth
		}
	}
	return result
}

func validateHierarchyContainment(t *testing.T, h *MeshHierarchy) {
	for _, c := range h.Children {
		for _, v := range c.Mesh.VertexSlice() {
			if !h.MeshSolid.Contains(v) {
				t.Fatal("child not contained in parent")
			}
		}
		validateHierarchyContainment(t, c)
	}
}

func BenchmarkMeshHierarchy(b *testing.B) {
	mesh, _, _ := hierarchyTestingMesh(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MeshToHierarchy(mesh)
	}
}

func hierarchyTestingMesh(f Failer) (mesh *Mesh, numHier int, depth int) {
	// This code created the original mesh:
	//
	// createShell := func(center Coord3D, rad float64) Solid {
	// 	return &SubtractedSolid{
	// 		Positive: &Sphere{Center: center, Radius: rad},
	// 		Negative: &Sphere{Center: center, Radius: rad-0.05},
	// 	}
	// }
	// solid := JoinedSolid{
	// 	// First hierarchy, depth 5
	// 	createShell(XYZ(0, 0, 0), 1.0),
	// 	createShell(XYZ(0, 0, 0), 0.5),
	// 	&Sphere{Radius: 0.1},

	// 	// Second hierarchy, depth 2
	// 	createShell(XYZ(3, 0, 0), 1.0),

	// 	// Third hierarchy, depth 1
	// 	&Sphere{Center: Y(3.0), Radius: 1.0},
	// }
	// mesh = MarchingCubesSearch(solid, 0.03, 2)
	// mesh.SaveGroupedSTL("test_data/hierarchy_test.stl")

	r, err := os.Open("test_data/hierarchy_test.stl.gz")
	if err != nil {
		f.Fatal(err)
	}
	defer r.Close()
	rz, err := gzip.NewReader(r)
	if err != nil {
		f.Fatal(err)
	}
	tris, err := ReadSTL(rz)
	if err != nil {
		f.Fatal(err)
	}
	return NewMeshTriangles(tris), 3, 5
}
