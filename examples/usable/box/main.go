package main

import (
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

const (
	Width         = 3.0
	Height        = 1.0
	Depth         = 4.0
	SideThickness = 0.2
	TopSpacing    = 0.025

	ScrewHoleSize   = 0.16
	HolderThickness = 0.2

	HandleHeight = 0.7
)

func main() {
	log.Println("Creating body...")
	MakeBody()
	log.Println("Creating lid...")
	MakeLid()
	log.Println("Creating handle...")
	MakeHandle()
}

func MakeBody() {
	mesh := model3d.NewMesh()
	addRect := func(p1, p2, p3 model3d.Coord3D) {
		p4 := p1.Add(p2.Sub(p1)).Add(p3.Sub(p1))
		mesh.Add(&model3d.Triangle{p1, p2, p3})
		mesh.Add(&model3d.Triangle{p2, p3, p4})
	}

	// Inside bottom
	addRect(
		model3d.Coord3D{X: SideThickness, Y: SideThickness, Z: SideThickness},
		model3d.Coord3D{X: Width + SideThickness, Y: SideThickness, Z: SideThickness},
		model3d.Coord3D{X: SideThickness, Y: Depth + SideThickness, Z: SideThickness},
	)

	// Outside bottom
	addRect(
		model3d.Coord3D{X: 0, Y: 0, Z: 0},
		model3d.Coord3D{X: Width + SideThickness*2, Y: 0, Z: 0},
		model3d.Coord3D{X: 0, Y: Depth + SideThickness*2, Z: 0},
	)

	// Left side.
	addRect(
		model3d.Coord3D{X: 0, Y: 0, Z: 0},
		model3d.Coord3D{X: 0, Y: 0, Z: Height},
		model3d.Coord3D{X: 0, Y: Depth + SideThickness*2, Z: 0},
	)

	// Right side.
	addRect(
		model3d.Coord3D{X: Width + SideThickness*2, Y: 0, Z: 0},
		model3d.Coord3D{X: Width + SideThickness*2, Y: 0, Z: Height},
		model3d.Coord3D{X: Width + SideThickness*2, Y: Depth + SideThickness*2, Z: 0},
	)

	// Front side.
	addRect(
		model3d.Coord3D{X: 0, Y: 0, Z: 0},
		model3d.Coord3D{X: 0, Y: 0, Z: Height},
		model3d.Coord3D{X: Width + SideThickness*2, Y: 0, Z: 0},
	)

	// Back side.
	addRect(
		model3d.Coord3D{X: 0, Y: Depth + SideThickness*2, Z: 0},
		model3d.Coord3D{X: 0, Y: Depth + SideThickness*2, Z: Height},
		model3d.Coord3D{X: Width + SideThickness*2, Y: Depth + SideThickness*2, Z: 0},
	)

	// Create top edges.
	for _, xOffset := range []float64{0, Width + SideThickness} {
		for _, yOffset := range []float64{0, Depth + SideThickness} {
			addRect(
				model3d.Coord3D{X: xOffset, Y: yOffset, Z: Height},
				model3d.Coord3D{X: xOffset + SideThickness, Y: yOffset, Z: Height},
				model3d.Coord3D{X: xOffset, Y: yOffset + SideThickness, Z: Height},
			)
			if xOffset == 0 {
				addRect(
					model3d.Coord3D{X: SideThickness, Y: yOffset, Z: Height},
					model3d.Coord3D{X: Width + SideThickness, Y: yOffset, Z: Height},
					model3d.Coord3D{X: SideThickness, Y: yOffset + SideThickness, Z: Height},
				)
			}
		}
		addRect(
			model3d.Coord3D{X: xOffset, Y: SideThickness, Z: Height},
			model3d.Coord3D{X: xOffset + SideThickness, Y: SideThickness, Z: Height},
			model3d.Coord3D{X: xOffset, Y: Depth + SideThickness, Z: Height},
		)
	}

	// Inside walls.
	for _, xOffset := range []float64{SideThickness, Width + SideThickness} {
		addRect(
			model3d.Coord3D{X: xOffset, Y: SideThickness, Z: SideThickness},
			model3d.Coord3D{X: xOffset, Y: SideThickness, Z: Height},
			model3d.Coord3D{X: xOffset, Y: Depth + SideThickness, Z: SideThickness},
		)
	}
	for _, yOffset := range []float64{SideThickness, Depth + SideThickness} {
		addRect(
			model3d.Coord3D{X: SideThickness, Y: yOffset, Z: SideThickness},
			model3d.Coord3D{X: SideThickness, Y: yOffset, Z: Height},
			model3d.Coord3D{X: Width + SideThickness, Y: yOffset, Z: SideThickness},
		)
	}

	mesh, _ = mesh.RepairNormals(1e-8)
	mesh = ScaleUp(mesh)
	ioutil.WriteFile("body.stl", mesh.EncodeSTL(), 0755)
}

func MakeLid() {
	cx := Width/2 + SideThickness
	cy := Depth/2 + SideThickness
	solid := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Rect{
				MinVal: model3d.Coord3D{},
				MaxVal: model3d.Coord3D{
					X: Width + SideThickness*2,
					Y: Depth + SideThickness*2,
					Z: SideThickness,
				},
			},
			&model3d.Rect{
				MinVal: model3d.Coord3D{
					X: SideThickness + TopSpacing,
					Y: SideThickness + TopSpacing,
					Z: SideThickness,
				},
				MaxVal: model3d.Coord3D{
					X: Width + SideThickness - TopSpacing,
					Y: Depth + SideThickness - TopSpacing,
					Z: SideThickness * 2,
				},
			},
		},
		Negative: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{X: cx, Y: cy},
				P2:     model3d.Coord3D{X: cx, Y: cy, Z: SideThickness},
				Radius: ScrewHoleSize,
			},
			&model3d.Rect{
				MinVal: model3d.Coord3D{
					X: SideThickness + TopSpacing + HolderThickness,
					Y: SideThickness + TopSpacing + HolderThickness,
					Z: SideThickness,
				},
				MaxVal: model3d.Coord3D{
					X: Width + SideThickness - TopSpacing - HolderThickness,
					Y: Depth + SideThickness - TopSpacing - HolderThickness,
					Z: SideThickness * 2,
				},
			},
		},
	}
	mesh := model3d.MarchingCubesSearch(solid, 0.0125, 8)
	mesh = ScaleUp(mesh)
	ioutil.WriteFile("lid.stl", mesh.EncodeSTL(), 0755)
}

func MakeHandle() {
	screw := model3d.JoinedSolid{
		&model3d.Cylinder{
			P1:     model3d.Coord3D{},
			P2:     model3d.Coord3D{Z: 0.2},
			Radius: 0.25,
		},
		&toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: 0.2},
			P2:         model3d.Coord3D{Z: HandleHeight},
			Radius:     0.14,
			GrooveSize: 0.05,
		},
	}
	mesh := model3d.MarchingCubesSearch(screw, 0.004, 8)
	mesh = ScaleUp(mesh)
	ioutil.WriteFile("screw.stl", mesh.EncodeSTL(), 0755)

	handle := &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Coord3D{Z: 0.2},
				P2:     model3d.Coord3D{Z: HandleHeight},
				Radius: 0.4,
			},
			NewScrewBase(),
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: 0.2},
			P2:         model3d.Coord3D{Z: HandleHeight},
			Radius:     0.16,
			GrooveSize: 0.05,
		},
	}
	mesh = model3d.MarchingCubesSearch(handle, 0.004, 8)
	mesh = ScaleUp(mesh)
	ioutil.WriteFile("handle.stl", mesh.EncodeSTL(), 0755)
}

type ScrewBase struct {
	Img image.Image
}

func NewScrewBase() *ScrewBase {
	r, err := os.Open("heart.png")
	essentials.Must(err)
	defer r.Close()
	img, err := png.Decode(r)
	essentials.Must(err)
	return &ScrewBase{Img: img}
}

func (s *ScrewBase) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -0.8, Y: -0.8, Z: 0}
}

func (s *ScrewBase) Max() model3d.Coord3D {
	return model3d.Coord3D{X: 0.8, Y: 0.8, Z: 0.2}
}

func (s *ScrewBase) Contains(c model3d.Coord3D) bool {
	if c.Z < 0 || c.Z > 0.2 {
		return false
	}
	rel := c.Sub(s.Min())
	rel.X /= s.Max().Sub(s.Min()).X
	rel.Y /= s.Max().Sub(s.Min()).Y
	if rel.X < 0 || rel.Y < 0 || rel.X > 1 || rel.Y > 1 {
		return false
	}

	r, _, _, a := s.Img.At(int(rel.X*160), int(rel.Y*160)).RGBA()

	if a < 0xffff/2 {
		return false
	}

	return r > 0xffff/2 || c.Z > 0.05
}

func ScaleUp(m *model3d.Mesh) *model3d.Mesh {
	return m.MapCoords(func(c model3d.Coord3D) model3d.Coord3D {
		return c.Scale(25.4)
	})
}
