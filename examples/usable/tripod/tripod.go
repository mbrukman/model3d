package main

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
	"github.com/unixpickle/model3d/toolbox3d"
)

func CreateTripod() model3d.Solid {
	return toolbox3d.ClampZMin(
		model3d.JoinedSolid{
			createLeg(0),
			createLeg(2 * math.Pi / 3.0),
			createLeg(4 * math.Pi / 3.0),
			model3d.StackSolids(
				&model3d.Cylinder{
					P1:     model3d.Z(TripodHeadZ),
					P2:     model3d.Z(TripodHeadZ + TripodHeadHeight),
					Radius: TripodHeadRadius,
				},
				&toolbox3d.ScrewSolid{
					P2:         model3d.Z(CradleBottomThickness - ScrewSlack),
					Radius:     ScrewRadius,
					GrooveSize: ScrewGroove,
				},
			),
		},
		0,
	)
}

func createLeg(theta float64) model3d.Solid {
	legEnd := model3d.Coord3D{X: TripodLegSpanRadius * math.Cos(theta), Y: TripodLegSpanRadius * math.Sin(theta)}
	footEnd := legEnd.Scale((TripodFootOutset + TripodLegSpanRadius) / TripodLegSpanRadius)
	return &model3d.SubtractedSolid{
		Positive: model3d.JoinedSolid{
			&model3d.Cylinder{
				P1:     model3d.Z(TripodHeight),
				P2:     legEnd,
				Radius: TripodLegRadius,
			},
			&model3d.Cylinder{
				P1:     footEnd,
				P2:     footEnd.Add(model3d.Z(TripodFootHeight)),
				Radius: TripodFootRadius,
			},
		},
		Negative: &toolbox3d.ScrewSolid{
			P1:         footEnd.Add(model3d.Coord3D{Z: -1e-5}),
			P2:         footEnd.Add(model3d.Z(1000)),
			Radius:     ScrewRadius + ScrewSlack,
			GrooveSize: ScrewGroove,
		},
	}
}
