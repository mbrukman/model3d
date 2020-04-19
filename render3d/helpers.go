package render3d

import (
	"math"

	"github.com/unixpickle/model3d/model3d"
)

const (
	helperFieldOfView = math.Pi / 3.6
	helperAmbient     = 0.1
	helperDiffuse     = 0.8
	helperSpecular    = 0.2
)

// ColorFunc determines a color for collisions on a
// surface. It is used for convenience methods where
// specifying a material would be too cumbersome.
type ColorFunc func(c model3d.Coord3D, rc model3d.RayCollision) Color

// TriangleColorFunc creates a ColorFunc from a function
// that colors triangles.
// This can be used for compatibility with
// model3d.EncodeMaterialOBJ.
//
// This only works when rendering meshes or triangles.
func TriangleColorFunc(f func(t *model3d.Triangle) [3]float64) ColorFunc {
	return func(_ model3d.Coord3D, rc model3d.RayCollision) Color {
		c := f(rc.Extra.(*model3d.TriangleCollision).Triangle)
		return NewColorRGB(c[0], c[1], c[2])
	}
}

type colorFuncObject struct {
	Object
	ColorFunc ColorFunc
}

func (c *colorFuncObject) Cast(r *model3d.Ray) (model3d.RayCollision, Material, bool) {
	rc, mat, ok := c.Object.Cast(r)
	if ok && c.ColorFunc != nil {
		p := r.Origin.Add(r.Direction.Scale(rc.Scale))
		color := c.ColorFunc(p, rc)
		mat = &PhongMaterial{
			Alpha:         10,
			SpecularColor: NewColor(helperSpecular),
			DiffuseColor:  color.Scale(helperDiffuse),
			AmbientColor:  color.Scale(helperAmbient),
		}
	}
	return rc, mat, ok
}

// Objectify turns a mesh, collider, or Object into a new
// Object with a specified coloration.
//
// Accepted object types are:
//
//     - render3d.Object
//     - *model3d.Mesh
//     - model3d.Collider
//
// The colorFunc is used to color the object's material.
// If colorFunc is used, a default yellow color is used,
// unless the object already has an associated material.
func Objectify(obj interface{}, colorFunc ColorFunc) Object {
	switch obj := obj.(type) {
	case Object:
		return &colorFuncObject{Object: obj, ColorFunc: colorFunc}
	case model3d.Collider:
		return &colorFuncObject{
			Object: &ColliderObject{
				Collider: obj,
				Material: &PhongMaterial{
					Alpha:         10,
					SpecularColor: NewColor(helperSpecular),
					DiffuseColor:  NewColorRGB(224.0/255, 209.0/255, 0).Scale(helperDiffuse),
					AmbientColor:  NewColorRGB(224.0/255, 209.0/255, 0).Scale(helperAmbient),
				},
			},
			ColorFunc: colorFunc,
		}
	case *model3d.Mesh:
		return Objectify(model3d.MeshToCollider(obj), colorFunc)
	default:
		panic("type not recognized")
	}
}

// SaveRendering renders a 3D object from the given point
// and saves the image to a file.
//
// The camera will automatically face the center of the
// object's bounding box.
//
// The obj argument must be supported by Objectify.
//
// If colorFunc is non-nil, it is used to determine the
// color for the visible parts of the model.
func SaveRendering(path string, obj interface{}, origin model3d.Coord3D, width, height int,
	colorFunc ColorFunc) error {
	object := Objectify(obj, colorFunc)
	image := NewImage(width, height)

	min, max := object.Min(), object.Max()
	center := min.Mid(max)
	caster := RayCaster{
		Camera: NewCameraAt(origin, center, helperFieldOfView),
		Lights: []*PointLight{
			&PointLight{
				Origin: center.Add(origin.Sub(center).Scale(1000)),
				Color:  NewColor(1.0),
			},
		},
	}
	caster.Render(image, object)
	return image.Save(path)
}

// SaveRandomGrid renders a 3D object from a variety of
// randomized angles and saves the grid of renderings to a
// file.
//
// The obj argument must be supported by Objectify.
//
// If colorFunc is non-nil, it is used to determine the
// color for the visible parts of the model.
func SaveRandomGrid(path string, obj interface{}, rows, cols, imgSize int,
	colorFunc ColorFunc) error {
	object := Objectify(obj, colorFunc)
	fullOutput := NewImage(cols*imgSize, rows*imgSize)

	min, max := object.Min(), object.Max()
	center := min.Mid(max)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			direction := model3d.NewCoord3DRandUnit()
			caster := &RayCaster{
				Camera: directionalCamera(object, direction),
				Lights: []*PointLight{
					&PointLight{
						Origin: center.Add(direction.Scale(1000)),
						Color:  NewColor(1.0),
					},
				},
			}
			subImage := NewImage(imgSize, imgSize)
			caster.Render(subImage, object)
			fullOutput.CopyFrom(subImage, j*imgSize, i*imgSize)
		}
	}

	return fullOutput.Save(path)
}

// directionalCamera figures out where to move a camera in
// the given unit direction to capture the bounding box of
// an object.
func directionalCamera(object Object, direction model3d.Coord3D) *Camera {
	min, max := object.Min(), object.Max()
	baseline := min.Dist(max)
	center := min.Mid(max)

	margin := 0.05
	minDist := baseline * 1e-4
	maxDist := baseline * 1e4
	for i := 0; i < 32; i++ {
		d := (minDist + maxDist) / 2
		cam := NewCameraAt(center.Add(direction.Scale(d)), center, helperFieldOfView)
		uncaster := cam.Uncaster(1, 1)
		contained := true
		for _, x := range []float64{min.X, max.X} {
			for _, y := range []float64{min.Y, max.Y} {
				for _, z := range []float64{min.Z, max.Z} {
					sx, sy := uncaster(model3d.Coord3D{X: x, Y: y, Z: z})
					if sx < margin || sy < margin || sx >= 1-margin || sy >= 1-margin {
						contained = false
					}
				}
			}
		}
		if contained {
			maxDist = d
		} else {
			minDist = d
		}
	}

	return NewCameraAt(center.Add(direction.Scale(maxDist)), center, helperFieldOfView)
}
