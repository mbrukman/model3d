package main

import (
	"image/png"
	"math"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

type Globe struct {
	Image *toolbox3d.Equirect
	Base  *render3d.Sphere
}

func NewGlobe() *Globe {
	r, err := os.Open("../../decoration/globe/map.png")
	essentials.Must(err)
	defer r.Close()
	mapImage, err := png.Decode(r)
	essentials.Must(err)

	return &Globe{
		Image: toolbox3d.NewEquirect(mapImage),
		Base: &render3d.Sphere{
			Center: model3d.Coord3D{Z: 0.5},
			Radius: 1.5,
		},
	}
}

func (g *Globe) Cast(r *model3d.Ray) (model3d.RayCollision, render3d.Material, bool) {
	collision, material, ok := g.Base.Cast(r)
	if !ok {
		return collision, material, ok
	}
	point := r.Origin.Add(r.Direction.Scale(collision.Scale)).Sub(g.Base.Center)

	point = model3d.NewMatrix3Rotation(model3d.Coord3D{Z: 1}, -math.Pi/2).MulColumn(point)
	point.Y, point.Z = point.Z, -point.Y

	red, green, blue, _ := g.Image.At(point.Geo()).RGBA()
	material = &render3d.PhongMaterial{
		Alpha:         5,
		SpecularColor: render3d.Color{X: 0.1, Y: 0.1, Z: 0.1},
		DiffuseColor: render3d.Color{X: float64(red) / 0xffff, Y: float64(green) / 0xffff,
			Z: float64(blue) / 0xffff}.Scale(0.9),
	}
	return collision, material, ok
}
