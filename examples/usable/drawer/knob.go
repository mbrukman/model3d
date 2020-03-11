package main

import (
	"github.com/unixpickle/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateKnob() model3d.Solid {
	return model3d.StackedSolid{
		&model3d.CylinderSolid{
			P2:     model3d.Coord3D{Z: KnobLength},
			Radius: KnobRadius,
		},
		&toolbox3d.ScrewSolid{
			P2:         model3d.Coord3D{Z: DrawerThickness + KnobNutThickness},
			GrooveSize: KnobScrewGroove,
			Radius:     KnobScrewRadius,
		},
	}
}

func CreateKnobNut() model3d.Solid {
	return &model3d.SubtractedSolid{
		Positive: &model3d.CylinderSolid{
			P2:     model3d.Coord3D{Z: KnobNutThickness},
			Radius: KnobNutThickness,
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         model3d.Coord3D{Z: -1e-5},
			P2:         model3d.Coord3D{Z: KnobNutThickness + 1e-5},
			GrooveSize: KnobScrewGroove,
			Radius:     KnobScrewRadius + KnobScrewSlack,
		},
	}
}
