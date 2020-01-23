package main

import (
	"log"

	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/model2d"
)

const (
	SideSize         = 7.0
	BottomThickness  = 0.2
	WallHeight       = 2.0
	WallThickness    = 0.3
	SectionHeight    = 1.7
	SectionThickness = 0.2

	LidThickness       = 0.3
	LidHolderThickness = 0.2
	LidHolderInset     = 0.06

	ImageSize = 1024.0
)

func main() {
	outline := model2d.MeshToCollider(
		model2d.MustReadBitmap("outline.png", nil).FlipY().Mesh(),
	)
	sections := model2d.MeshToCollider(
		model2d.MustReadBitmap("sections.png", nil).FlipY().Mesh(),
	)

	log.Println("Creating box...")
	mesh := model3d.SolidToMesh(&BoxSolid{Outline: outline, Sections: sections}, 0.02, 0, -1, 10)
	log.Println(" - flattening base...")
	mesh = mesh.FlattenBase(0)
	log.Println(" - saving...")
	mesh.SaveGroupedSTL("box.stl")
	log.Println(" - rendering...")
	model3d.SaveRandomGrid("rendering_box.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)

	log.Println("Creating lid...")
	mesh = model3d.SolidToMesh(&LidSolid{Outline: outline}, 0.02, 0, -1, 20)
	log.Println(" - flattening base...")
	mesh = mesh.FlattenBase(0)
	log.Println(" - saving...")
	mesh.SaveGroupedSTL("lid.stl")
	log.Println(" - rendering...")
	model3d.SaveRandomGrid("rendering_lid.png", model3d.MeshToCollider(mesh), 3, 3, 300, 300)
}

type BoxSolid struct {
	Outline  model2d.Collider
	Sections model2d.Collider
}

func (b *BoxSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (b *BoxSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: WallHeight}
}

func (b *BoxSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(b, c) {
		return false
	}
	scale := ImageSize / SideSize
	c2 := c.Coord2D().Scale(scale)
	if !model2d.ColliderContains(b.Outline, c2, 0) {
		return false
	}
	if c.Z < BottomThickness {
		return true
	}
	if b.Outline.CircleCollision(c2, WallThickness*scale) {
		return true
	}
	return c.Z < SectionHeight && b.Sections.CircleCollision(c2, 0.5*SectionThickness*scale)
}

type LidSolid struct {
	Outline model2d.Collider
}

func (l *LidSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{}
}

func (l *LidSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: SideSize, Y: SideSize, Z: LidThickness + LidHolderThickness}
}

func (l *LidSolid) Contains(c model3d.Coord3D) bool {
	if !model3d.InSolidBounds(l, c) {
		return false
	}
	scale := ImageSize / SideSize
	c2 := c.Coord2D().Scale(scale)
	if c.Z < LidThickness {
		return model2d.ColliderContains(l.Outline, c2, 0)
	} else {
		return model2d.ColliderContains(l.Outline, c2, scale*(WallThickness+LidHolderInset))
	}
}
