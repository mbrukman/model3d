package main

import (
	"log"
	"math"

	"github.com/unixpickle/model3d"
)

const (
	NumSides   = 12
	BaseHeight = 0.4
	TipHeight  = 1.2
)

func main() {
	log.Println("Creating diamond polytope...")
	system := CreateDiamondPolytope()
	log.Println("Exporting diamond...")
	mesh := system.Mesh()
	mesh.SaveGroupedSTL("diamond.stl")
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	CreateStand(mesh)
}

func CreateDiamondPolytope() model3d.ConvexPolytope {
	system := model3d.ConvexPolytope{
		&model3d.LinearConstraint{
			Normal: model3d.Coord3D{Z: -1},
			Max:    BaseHeight,
		},
	}

	addTriangle := func(t *model3d.Triangle) {
		n := t.Normal()

		// Make sure the normal points outward.
		if n.Dot(t[0]) < 0 {
			t[0], t[1] = t[1], t[0]
		}

		system = append(system, &model3d.LinearConstraint{
			Normal: t.Normal(),
			Max:    t[0].Dot(t.Normal()),
		})
	}

	iAngle := math.Pi * 2 / NumSides
	rimPoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos(float64(i) * iAngle),
			Y: math.Sin(float64(i) * iAngle),
		}
	}
	basePoint := func(i int) model3d.Coord3D {
		return model3d.Coord3D{
			X: math.Cos((float64(i) + 0.5) * iAngle),
			Y: math.Sin((float64(i) + 0.5) * iAngle),
		}.Scale(1 - BaseHeight).Sub(model3d.Coord3D{Z: BaseHeight})
	}
	tipPoint := model3d.Coord3D{Z: TipHeight}

	for i := 0; i < NumSides; i++ {
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			tipPoint,
		})
		addTriangle(&model3d.Triangle{
			rimPoint(i),
			rimPoint(i + 1),
			basePoint(i),
		})
		addTriangle(&model3d.Triangle{
			basePoint(i),
			basePoint(i + 1),
			rimPoint(i + 1),
		})
	}
	return system
}

func CreateStand(diamond *model3d.Mesh) {
	log.Println("Creating stand...")
	diamond = diamond.MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: -1}.Mul)
	solid := model3d.NewColliderSolid(model3d.MeshToCollider(diamond))

	standSolid := &model3d.SubtractedSolid{
		Positive: &model3d.CylinderSolid{
			P1:     model3d.Coord3D{Z: solid.Min().Z},
			P2:     model3d.Coord3D{Z: solid.Min().Z + 0.5},
			Radius: 1.0,
		},
		Negative: solid,
	}
	mesh := model3d.SolidToMesh(standSolid, 0.01, 0, 0, 0)
	smoother := &model3d.MeshSmoother{
		StepSize:           0.1,
		Iterations:         200,
		ConstraintDistance: 0.01,
		ConstraintWeight:   0.1,
	}
	mesh = smoother.Smooth(mesh)
	mesh = mesh.FlattenBase(0)

	mesh.SaveGroupedSTL("stand.stl")
	model3d.SaveRandomGrid("rendering_stand.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}
