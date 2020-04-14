package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"

	"github.com/unixpickle/model3d/model2d"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/render3d"
)

const Thickness = 0.9

func main() {
	pumpkin := &model3d.SubtractedSolid{
		Positive: PumpkinSolid{Scale: 1},
		Negative: PumpkinSolid{Scale: Thickness},
	}
	base := &model3d.SubtractedSolid{
		Positive: LidSolid{Solid: pumpkin},
		Negative: NewEtchSolid(),
	}
	lid := LidSolid{IsLid: true, Solid: model3d.JoinedSolid{pumpkin, StemSolid{}}}

	colorFunc := func(t *model3d.Triangle) [3]float64 {
		c := t[0]
		if (PumpkinSolid{Scale: 1.025}).Contains(c) {
			expectedNormal := t[0].Geo().Coord3D()
			if math.Abs(expectedNormal.Dot(t.Normal())) > 0.5 {
				return [3]float64{214.0 / 255, 143.0 / 255, 0}
			} else {
				return [3]float64{255.0 / 255, 206.0 / 255, 107.0 / 255}
			}
		}
		return [3]float64{79.0 / 255, 53.0 / 255, 0}
	}

	outsideMesh := model3d.NewMesh()
	insideMesh := model3d.NewMesh()
	stemMesh := model3d.NewMesh()

	log.Println("Creating mesh...")
	mesh := model3d.MarchingCubesSearch(base, 0.02, 8)
	mesh.AddMesh(model3d.MarchingCubesSearch(lid, 0.02, 8))

	fmt.Println(mesh.Min(), mesh.Max())

	dec := &model3d.Decimator{
		FeatureAngle:       0.1,
		PlaneDistance:      4e-4,
		BoundaryDistance:   1e-5,
		MinimumAspectRatio: 0.01,
	}
	fmt.Println("before", len(mesh.TriangleSlice()))
	mesh = dec.Decimate(mesh)
	fmt.Println("after", len(mesh.TriangleSlice()))

	mesh.Iterate(func(t *model3d.Triangle) {
		c := t[0]
		if (PumpkinSolid{Scale: 1.025}).Contains(c) {
			expectedNormal := t[0].Geo().Coord3D()
			if math.Abs(expectedNormal.Dot(t.Normal())) > 0.5 {
				outsideMesh.Add(t)
			} else {
				insideMesh.Add(t)
			}
		} else {
			stemMesh.Add(t)
		}
	})

	mat := model3d.NewMatrix3Rotation(
		model3d.Coord3D{X: 1}, math.Pi/2,
	)
	mat = model3d.NewMatrix3Rotation(
		model3d.Coord3D{Z: 1}, -math.Pi/2,
	).Mul(mat)
	outsideMesh = outsideMesh.MapCoords(mat.MulColumn)
	insideMesh = insideMesh.MapCoords(mat.MulColumn)
	stemMesh = stemMesh.MapCoords(mat.MulColumn)

	outsideMesh.SaveGroupedSTL("pumpkin_outside.stl")
	insideMesh.SaveGroupedSTL("pumpkin_inside.stl")
	stemMesh.SaveGroupedSTL("pumpkin_stem.stl")

	log.Println("Saving mesh...")
	ioutil.WriteFile("pumpkin.zip", mesh.EncodeMaterialOBJ(colorFunc), 0755)

	log.Println("Rendering...")
	render3d.SaveRandomGrid("rendering.png", mesh, 3, 3, 300, render3d.TriangleColorFunc(colorFunc))
}

type PumpkinSolid struct {
	Scale float64
}

func (p PumpkinSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -p.Scale * 1.6, Y: -p.Scale * 1.6, Z: -p.Scale * 1.6}
}

func (p PumpkinSolid) Max() model3d.Coord3D {
	return p.Min().Scale(-1)
}

func (p PumpkinSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(p, c) {
		return false
	}
	g := c.Geo()
	r := p.Scale * (1 + 0.1*math.Abs(math.Sin(g.Lon*4)) + 0.5*math.Cos(g.Lat))
	return c.Norm() <= r
}

type StemSolid struct{}

func (s StemSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{Y: 0.9, X: -0.3, Z: -0.3}
}

func (s StemSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{Y: 1.6, X: 0.3, Z: 0.3}
}

func (s StemSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(s, c) {
		return false
	}
	c.X -= 0.15 * math.Pow(c.Y-s.Min().Y, 2)
	theta := math.Atan2(c.X, c.Z)
	radius := 0.05*math.Sin(theta*5) + 0.15
	return model3d.Coord2D{X: c.X, Y: c.Z}.Norm() < radius
}

type LidSolid struct {
	IsLid bool
	Solid model3d.Solid
}

func (l LidSolid) Min() model3d.Coord3D {
	return l.Solid.Min()
}

func (l LidSolid) Max() model3d.Coord3D {
	return l.Solid.Max()
}

func (l LidSolid) Contains(c model3d.Coord3D) bool {
	coneCenter := 0.0
	if l.IsLid {
		coneCenter += 0.1
	}
	inLid := model3d.Coord2D{X: c.X, Y: c.Z}.Norm() < 0.7*(c.Y-coneCenter)
	return inLid == l.IsLid && l.Solid.Contains(c)
}

type EtchSolid struct {
	Solid  model2d.Solid
	Radius float64
	Height float64
}

func NewEtchSolid() *EtchSolid {
	bmp := model2d.MustReadBitmap("etching.png", nil)
	scale := model2d.Coord{X: 1 / float64(bmp.Width), Y: 1 / float64(bmp.Height)}
	mesh := bmp.Mesh().SmoothSq(50).MapCoords(scale.Mul)
	return &EtchSolid{
		Solid:  model2d.NewColliderSolid(model2d.MeshToCollider(mesh)),
		Radius: 1.6,
		Height: 1.5,
	}
}

func (e *EtchSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -e.Radius, Y: -e.Height, Z: -e.Radius}
}

func (e *EtchSolid) Max() model3d.Coord3D {
	return e.Min().Scale(-1)
}

func (e *EtchSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InBounds(e, c) {
		return false
	}
	xFrac := c.Geo().Lon/(math.Pi*2) + 0.5
	yFrac := 1 - (c.Y+e.Height)/(e.Height*2)
	return e.Solid.Contains(model2d.Coord{X: xFrac, Y: yFrac})
}
