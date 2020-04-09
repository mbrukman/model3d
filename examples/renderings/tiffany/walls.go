package main

import (
	"math"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

var TiffanyBlue = render3d.NewColorRGB(129.0/255.0, 216.0/255.0, 208.0/255.0)

type Walls struct {
	Base   *render3d.ColliderObject
	Lights []*CeilingLight
}

func NewWalls() *Walls {
	mesh := model3d.NewMeshRect(model3d.Coord3D{X: -6, Y: -15, Z: -5},
		model3d.Coord3D{X: 6, Y: 8, Z: 5})

	// Face normals inward.
	mesh = mesh.MapCoords(model3d.Coord3D{X: -1, Y: 1, Z: 1}.Mul)

	var lights []*CeilingLight
	for x := -3; x <= 3; x += 3 {
		for y := -5; y <= 4; y += 3 {
			lights = append(lights, &CeilingLight{
				Cylinder: &model3d.Cylinder{
					P1: model3d.Coord3D{X: float64(x), Y: float64(y), Z: mesh.Max().Z},
					P2: model3d.Coord3D{X: float64(x), Y: float64(y),
						Z: mesh.Max().Z + CeilingLightDepth},
					Radius: CeilingLightRadius,
				},
			})
		}
	}

	return &Walls{
		Base: &render3d.ColliderObject{
			Collider: model3d.MeshToCollider(mesh),
			Material: &render3d.LambertMaterial{
				DiffuseColor: render3d.NewColor(1),
			},
		},
		Lights: lights,
	}
}

func (w *Walls) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	collision, material, ok := w.Base.Cast(r)
	if !ok || math.Abs(collision.Normal.Z+1) < 1e-8 {
		for _, l := range w.Lights {
			if ok && l.Cut(r, collision) {
				ok = false
			}
			lc, lmat, lok := l.Cast(r)
			if !lok {
				continue
			}
			if !ok || lc.Scale < collision.Scale {
				ok = true
				collision = lc
				material = lmat
			}
		}
	}
	if !ok {
		return collision, material, ok
	}
	point := r.Origin.Add(r.Direction.Scale(collision.Scale))
	c2d := point.Coord2D()
	c2d.Y += 2

	if math.Abs(point.X-w.Base.Collider.Max().X) < 1e-8 {
		// Striped wall
		scale := 1.0
		if math.Abs(math.Mod(point.Z+point.Y+100, 1.0)-0.5) < 0.25 {
			scale = 1.2
		}
		material = &render3d.LambertMaterial{
			DiffuseColor: TiffanyBlue.Scale(scale),
		}
	} else if math.Abs(point.X-w.Base.Collider.Min().X) < 1e-8 {
		// Dotted wall
		scale := 1.2
		zMod1 := math.Mod(point.Z+100, 2) - 1
		yMod1 := math.Mod(point.Y+100, 4) - 2
		zMod2 := math.Mod(point.Z+101, 2) - 1
		yMod2 := math.Mod(point.Y+102, 4) - 2
		if math.Sqrt(zMod1*zMod1+yMod1*yMod1) < 0.4 || math.Sqrt(zMod2*zMod2+yMod2*yMod2) < 0.4 {
			scale = 1.0
		}
		material = &render3d.LambertMaterial{
			DiffuseColor: TiffanyBlue.Scale(scale),
		}
	} else if math.Abs(point.Y-w.Base.Collider.Max().Y) < 1e-8 {
		// Gray back wall pattern.
		scale := 1.0
		if math.Abs(math.Mod(point.Z+math.Pow(math.Sin(point.X*3), 2)/3+100, 1.0/3)-0.5/3) < 0.25/3 {
			scale = 0.9
		}
		material = &render3d.LambertMaterial{
			DiffuseColor: render3d.NewColor(scale),
		}
	} else if math.Abs(point.Z-w.Base.Collider.Min().Z) < 1e-8 {
		// Checkerboard floor.
		diffuse := 0.0
		xMod := math.Mod(point.X+100, 2)
		yMod := math.Mod(point.Y+100, 2)
		if (xMod < 1) == (yMod < 1) {
			diffuse = 0.9
		}
		material = &render3d.PhongMaterial{
			Alpha:         50,
			SpecularColor: render3d.NewColor(0.1),
			DiffuseColor:  render3d.NewColor(diffuse),
		}
	}

	return collision, material, ok
}
