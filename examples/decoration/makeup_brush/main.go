package main

import (
	"io/ioutil"
	"math"

	"github.com/unixpickle/model3d"
)

const (
	HandleHeight       = 0.7
	HandleBottomRadius = 0.1
	HandleTopRadius    = 0.2

	BrushHeight         = 0.3
	BrushTopRadius      = 0.4
	BrushRippleDepth    = 0.02
	BrushTopRippleDepth = 0.01
	BrushRippleFreq     = 25.0
	BrushTopRippleFreq  = 40.0
)

func main() {
	solid := model3d.JoinedSolid{
		BrushSolid{},
		// Rounded bottom for brush.
		&model3d.SphereSolid{
			Center: model3d.Coord3D{Z: -HandleHeight},
			Radius: HandleBottomRadius,
		},
	}
	mesh := model3d.SolidToMesh(solid, 0.01, 2, -1, 5)
	ioutil.WriteFile("brush.stl", mesh.EncodeSTL(), 0755)
	model3d.SaveRandomGrid("rendering.png", model3d.MeshToCollider(mesh), 3, 3, 200, 200)
}

type BrushSolid struct{}

func (b BrushSolid) Min() model3d.Coord3D {
	return model3d.Coord3D{X: -BrushTopRadius, Y: -BrushTopRadius, Z: -HandleHeight}
}

func (b BrushSolid) Max() model3d.Coord3D {
	return model3d.Coord3D{X: BrushTopRadius, Y: BrushTopRadius, Z: BrushHeight + BrushTopRadius/2}
}

func (b BrushSolid) Contains(c model3d.Coord3D) bool {
	if c.Min(b.Min()) != b.Min() || c.Max(b.Max()) != b.Max() {
		return false
	}

	centerDist := (model3d.Coord2D{X: c.X, Y: c.Y}).Norm()

	if c.Z < 0 {
		// Handle
		frac := (c.Z + HandleHeight) / HandleHeight
		radius := frac*HandleTopRadius + (1-frac)*HandleBottomRadius
		return centerDist <= radius
	}

	frac := math.Max(0, (BrushHeight-c.Z)/BrushHeight)
	radius := frac*HandleTopRadius + (1-frac)*BrushTopRadius - BrushRippleDepth
	theta := math.Atan2(c.Y, c.X)
	radius -= BrushRippleDepth * math.Sin(theta*BrushRippleFreq)
	if c.Z < BrushHeight {
		return centerDist <= radius
	}

	// Top of the brush.
	topRadius := BrushTopRadius - BrushRippleDepth*(1+math.Sin(theta*BrushRippleFreq))
	if centerDist >= topRadius {
		return false
	}
	height := math.Sqrt(1-math.Pow((10+centerDist/topRadius)/11, 2)) * topRadius
	height += BrushTopRippleDepth * math.Abs(math.Cos((centerDist/topRadius)*BrushTopRippleFreq))
	return c.Z-BrushHeight < height
}
