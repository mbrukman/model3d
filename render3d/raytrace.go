package render3d

import (
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d"
)

const DefaultEpsilon = 1e-8

// A RecursiveRayTracer renders objects using recursive
// tracing with random sampling.
type RecursiveRayTracer struct {
	Camera *Camera
	Lights []*PointLight

	// FocusPoints are functions which cause rays to
	// bounce more in certain directions, with the aim of
	// reducing variance with no bias.
	FocusPoints []FocusPoint

	// FocusPointProbs stores, for each FocusPoint, the
	// probability that this focus point is used to sample
	// a ray (rather than the BRDF).
	FocusPointProbs []float64

	// MaxDepth is the maximum number of recursions.
	// Setting to 0 is almost equivalent to RayCast, but
	// the ray tracer still checks for shadows.
	MaxDepth int

	// NumSamples is the number of rays to sample.
	NumSamples int

	// Cutoff is the maximum brightness for which
	// recursion is performed. If small but non-zero, the
	// number of rays traced can be reduced.
	Cutoff float64

	// Antialias, if non-zero, specifies a fraction of a
	// pixel to perturb every ray's origin.
	// Thus, 1 is maximum, and 0 means no change.
	Antialias float64

	// Epsilon is a small distance used to move away from
	// surfaces before bouncing new rays.
	// If nil, DefaultEpsilon is used.
	Epsilon float64

	// LogFunc, if specified, is called periodically with
	// progress information.
	LogFunc func(frac float64)
}

// Render renders the object to an image.
func (r *RecursiveRayTracer) Render(img *Image, obj Object) {
	if r.NumSamples == 0 {
		panic("must set NumSamples to non-zero for RecursiveRayTracer")
	}
	maxX := float64(img.Width) - 1
	maxY := float64(img.Height) - 1
	caster := r.Camera.Caster(maxX, maxY)

	coords := make(chan [3]int, img.Width*img.Height)
	var idx int
	for y := 0; y < img.Width; y++ {
		for x := 0; x < img.Height; x++ {
			coords <- [3]int{x, y, idx}
			idx++
		}
	}
	close(coords)

	progressCh := make(chan struct{}, 1)

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gen := rand.New(rand.NewSource(rand.Int63()))
			ray := model3d.Ray{Origin: r.Camera.Origin}
			for c := range coords {
				ray.Direction = caster(float64(c[0]), float64(c[1]))
				var color Color
				for i := 0; i < r.NumSamples; i++ {
					if r.Antialias != 0 {
						dx := gen.Float64() - 0.5
						dy := gen.Float64() - 0.5
						ray.Direction = caster(float64(c[0])+dx, float64(c[1])+dy)
					}
					color = color.Add(r.recurse(gen, obj, &ray, 0, Color{X: 1, Y: 1, Z: 1}))
				}
				img.Data[c[2]] = color.Scale(1 / float64(r.NumSamples))
				progressCh <- struct{}{}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(progressCh)
	}()

	updateInterval := essentials.MaxInt(1, img.Width*img.Height/1000)
	var pixelsComplete int
	for _ = range progressCh {
		if r.LogFunc != nil {
			pixelsComplete++
			if pixelsComplete%updateInterval == 0 {
				r.LogFunc(float64(pixelsComplete) / float64(img.Width*img.Height))
			}
		}
	}
}

func (r *RecursiveRayTracer) recurse(gen *rand.Rand, obj Object, ray *model3d.Ray,
	depth int, scale Color) Color {
	if scale.Sum()/3 < r.Cutoff {
		return Color{}
	}
	collision, material, ok := obj.Cast(ray)
	if !ok {
		return Color{}
	}
	point := ray.Origin.Add(ray.Direction.Scale(collision.Scale))

	dest := ray.Direction.Normalize().Scale(-1)
	color := material.Emission()
	if depth == 0 {
		// Only add ambient light directly to object, not to
		// recursive rays.
		color = color.Add(material.Ambient())
	}
	for _, l := range r.Lights {
		lightDirection := l.Origin.Sub(point)

		shadowRay := r.bounceRay(point, lightDirection)
		collision, _, ok := obj.Cast(shadowRay)
		if ok && collision.Scale < 1 {
			continue
		}

		brdf := material.BSDF(collision.Normal, point.Sub(l.Origin).Normalize(), dest)
		color = color.Add(l.ShadeCollision(collision.Normal, lightDirection).Mul(brdf))
	}
	if depth >= r.MaxDepth {
		return color
	}
	nextSource := r.sampleNextSource(gen, point, collision.Normal, dest, material)
	weight := 1 / r.sourceDensity(point, collision.Normal, nextSource, dest, material)
	weight *= math.Abs(nextSource.Dot(collision.Normal))
	reflectWeight := material.BSDF(collision.Normal, nextSource, dest)
	nextRay := r.bounceRay(point, nextSource.Scale(-1))
	nextMask := reflectWeight.Scale(weight)
	nextScale := scale.Mul(nextMask)
	nextColor := r.recurse(gen, obj, nextRay, depth+1, nextScale)
	return color.Add(nextColor.Mul(nextMask))
}

func (r *RecursiveRayTracer) sampleNextSource(gen *rand.Rand, point, normal, dest model3d.Coord3D,
	mat Material) model3d.Coord3D {
	if len(r.FocusPoints) == 0 {
		return mat.SampleSource(gen, normal, dest)
	}

	p := rand.Float64()
	for i, prob := range r.FocusPointProbs {
		p -= prob
		if p < 0 {
			return r.FocusPoints[i].SampleFocus(gen, point)
		}
	}

	return mat.SampleSource(gen, normal, dest)
}

func (r *RecursiveRayTracer) sourceDensity(point, normal, source, dest model3d.Coord3D,
	mat Material) float64 {
	if len(r.FocusPoints) == 0 {
		return mat.SourceDensity(normal, source, dest)
	}

	matProb := 1.0
	var prob float64
	for i, focusProb := range r.FocusPointProbs {
		prob += focusProb * r.FocusPoints[i].FocusDensity(point, source)
		matProb -= focusProb
	}

	return prob + matProb*mat.SourceDensity(normal, source, dest)
}

func (r *RecursiveRayTracer) bounceRay(point model3d.Coord3D, dir model3d.Coord3D) *model3d.Ray {
	eps := r.Epsilon
	if eps == 0 {
		eps = DefaultEpsilon
	}
	return &model3d.Ray{
		// Prevent a duplicate collision from being
		// detected when bouncing off an existing
		// object.
		Origin:    point.Add(dir.Normalize().Scale(eps)),
		Direction: dir,
	}
}
