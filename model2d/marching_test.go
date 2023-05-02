package model2d

import (
	"math/rand"
	"testing"
)

func TestMarchingSquares(t *testing.T) {
	solid := BitmapToSolid(testingBitmap())

	testMesh := func(mesh *Mesh) {
		MustValidateMesh(t, mesh, true)

		meshSolid := NewColliderSolid(MeshToCollider(mesh))

		for i := 0; i < 1000; i++ {
			point := Coord{
				X: float64(rand.Intn(int(solid.Max().X) + 2)),
				Y: float64(rand.Intn(int(solid.Max().Y) + 2)),
			}
			if solid.Contains(point) != meshSolid.Contains(point) {
				t.Error("containment mismatch at:", point)
			}
		}
	}

	t.Run("Plain", func(t *testing.T) {
		mesh := MarchingSquares(solid, 1.0)
		testMesh(mesh)
	})

	t.Run("Search", func(t *testing.T) {
		mesh := MarchingSquaresSearch(solid, 1.0, 8)
		testMesh(mesh)
	})
}

func TestMarchingSquaresASCII(t *testing.T) {
	expected :=
		`                                                                ` + "\n" +
			`                 /\                          /\                 ` + "\n" +
			`        ________/  \________        ________/  \________        ` + "\n" +
			`       /                    \      /                    \       ` + "\n" +
			`      /                      \    /                      \      ` + "\n" +
			`     /                        \  /                        \     ` + "\n" +
			`    /                          \/                          \    ` + "\n" +
			`   /                                                        \   ` + "\n" +
			`  /                                                          \  ` + "\n" +
			` /                                                            \ ` + "\n" +
			` \                                                            / ` + "\n" +
			`  \                                                          /  ` + "\n" +
			`   \                                                        /   ` + "\n" +
			`    \                          /\                          /    ` + "\n" +
			`     \                        /  \                        /     ` + "\n" +
			`      \                      /    \                      /      ` + "\n" +
			`       \________    ________/      \________    ________/       ` + "\n" +
			`                \  /                        \  /                ` + "\n" +
			`                 \/                          \/                 ` + "\n" +
			`                                                                `
	solid := JoinedSolid{
		&Circle{Radius: 8.0},
		&Circle{Center: X(14), Radius: 8.0},
	}
	ascii := MarchingSquaresASCII(solid, 1.0)
	if ascii != expected {
		t.Errorf("expected:\n----\n%s\n----\nbut got:\n----\n%s\n----\n", expected, ascii)
	}
}

func TestMarchingSquaresFilter(t *testing.T) {
	t.Run("Circle", func(t *testing.T) {
		mesh := NewMeshPolar(func(t float64) float64 {
			return 1.0
		}, 400).Translate(XY(0.1, -0.2))
		collider := MeshToCollider(mesh)
		solid := NewColliderSolid(collider)
		base := MarchingSquares(solid, 0.1)
		rc := MarchingSquaresFilter(solid, collider.RectCollision, 0.1)
		if !meshesEqual(base, rc) {
			t.Fatal("meshes should be equal")
		}
	})
	t.Run("Boxes", func(t *testing.T) {
		mesh := NewMesh()
		mesh.AddMesh(NewMeshRect(XY(-1, -1), XY(0, 0)))
		mesh.AddMesh(NewMeshRect(XY(0.1, 0), XY(1, 1)))
		collider := MeshToCollider(mesh)
		solid := NewColliderSolid(collider)
		base := MarchingSquares(solid, 0.1)
		rc := MarchingSquaresFilter(solid, collider.RectCollision, 0.1)
		if !meshesEqual(base, rc) {
			t.Fatal("meshes should be equal")
		}
	})
}
