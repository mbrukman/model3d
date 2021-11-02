package model3d

import (
	"archive/zip"
	"bufio"
	"bytes"
	"io"
	"strconv"

	"github.com/pkg/errors"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/model3d/fileformats"
)

// EncodeSTL encodes a list of triangles in the binary STL
// format for use in 3D printing.
func EncodeSTL(triangles []*Triangle) []byte {
	var buf bytes.Buffer
	WriteSTL(&buf, triangles)
	return buf.Bytes()
}

// WriteSTL writes a list of triangles in the binary STL
// format to w.
func WriteSTL(w io.Writer, triangles []*Triangle) error {
	if err := writeSTL(w, triangles); err != nil {
		return errors.Wrap(err, "write STL")
	}
	return nil
}

func writeSTL(w io.Writer, triangles []*Triangle) error {
	if int(uint32(len(triangles))) != len(triangles) {
		return errors.New("too many triangles for STL format")
	}
	bw := bufio.NewWriter(w)
	writer, err := fileformats.NewSTLWriter(bw, uint32(len(triangles)))
	if err != nil {
		return err
	}

	for _, t := range triangles {
		verts := [3][3]float32{
			castVector32(t[0]),
			castVector32(t[1]),
			castVector32(t[2]),
		}
		if err := writer.WriteTriangle(castVector32(t.Normal()), verts); err != nil {
			return err
		}
	}
	return bw.Flush()
}

func castVector32(v Coord3D) [3]float32 {
	var res [3]float32
	for i, x := range v.Array() {
		res[i] = float32(x)
	}
	return res
}

// EncodePLY encodes a 3D model as a PLY file, including
// colors for every vertex.
//
// The colorFunc maps coordinates to 24-bit RGB colors.
func EncodePLY(triangles []*Triangle, colorFunc func(Coord3D) [3]uint8) []byte {
	var buf bytes.Buffer
	WritePLY(&buf, triangles, colorFunc)
	return buf.Bytes()
}

// WritePLY writes the 3D model as a PLY file, including
// colors for every vertex.
//
// The colorFunc maps coordinates to 24-bit RGB colors.
func WritePLY(w io.Writer, triangles []*Triangle, colorFunc func(Coord3D) [3]uint8) error {
	coords := [][3]float64{}
	colors := [][3]uint8{}
	coordToIdx := NewCoordToInt()
	for _, t := range triangles {
		for _, p := range t {
			if _, ok := coordToIdx.Load(p); !ok {
				coordToIdx.Store(p, len(coords))
				coords = append(coords, p.Array())
				colors = append(colors, colorFunc(p))
			}
		}
	}

	p, err := fileformats.NewPLYWriter(w, len(coords), len(triangles))
	if err != nil {
		return err
	}

	for i, c := range coords {
		if err := p.WriteCoord(c, colors[i]); err != nil {
			return err
		}
	}
	for _, t := range triangles {
		idxs := [3]int{
			coordToIdx.Value(t[0]),
			coordToIdx.Value(t[1]),
			coordToIdx.Value(t[2]),
		}
		if err := p.WriteTriangle(idxs); err != nil {
			return err
		}
	}

	return nil
}

// EncodeMaterialOBJ encodes a 3D model as a zip file
// containing both an OBJ and an MTL file.
//
// The colorFunc maps faces to real-valued RGB colors.
//
// The encoding creates a different material for every
// color, so the resulting file will be much smaller if a
// few identical colors are reused for many triangles.
func EncodeMaterialOBJ(triangles []*Triangle, colorFunc func(t *Triangle) [3]float64) []byte {
	var buf bytes.Buffer
	WriteMaterialOBJ(&buf, triangles, colorFunc)
	return buf.Bytes()
}

// WriteMaterialOBJ encodes a 3D model as a zip file
// containing both an OBJ and an MTL file.
//
// The colorFunc maps faces to real-valued RGB colors.
//
// The encoding creates a different material for every
// color, so the resulting file will be much smaller if a
// few identical colors are reused for many triangles.
func WriteMaterialOBJ(w io.Writer, ts []*Triangle, colorFunc func(t *Triangle) [3]float64) error {
	if err := writeMaterialOBJ(w, ts, colorFunc); err != nil {
		return errors.Wrap(err, "write material OBJ")
	}
	return nil
}

func writeMaterialOBJ(w io.Writer, triangles []*Triangle,
	colorFunc func(t *Triangle) [3]float64) error {
	obj, mtl := BuildMaterialOBJ(triangles, colorFunc)

	zipFile := zip.NewWriter(w)

	fw, err := zipFile.Create("object.obj")
	if err != nil {
		return err
	}
	if err := obj.Write(fw); err != nil {
		return err
	}

	fw, err = zipFile.Create("material.mtl")
	if err != nil {
		return err
	}
	if err := mtl.Write(fw); err != nil {
		return err
	}

	return zipFile.Close()
}

// BuildMaterialOBJ constructs obj and mtl files from a
// triangle mesh where each triangle's color is determined
// by a function c.
//
// Since the obj file must reference the mtl file, it does
// so by the name "material.mtl". Change o.MaterialFiles
// if this is not desired.
func BuildMaterialOBJ(t []*Triangle, c func(t *Triangle) [3]float64) (o *fileformats.OBJFile,
	m *fileformats.MTLFile) {
	o = &fileformats.OBJFile{
		MaterialFiles: []string{"material.mtl"},
	}
	m = &fileformats.MTLFile{}

	triColors := make([][3]float32, len(t))
	essentials.ConcurrentMap(0, len(t), func(i int) {
		tri := t[i]
		color64 := c(tri)
		triColors[i] = [3]float32{float32(color64[0]), float32(color64[1]), float32(color64[2])}
	})

	colorToMat := map[[3]float32]int{}
	coordToIdx := NewCoordToInt()
	for i, tri := range t {
		color32 := triColors[i]
		matIdx, ok := colorToMat[color32]
		var group *fileformats.OBJFileFaceGroup
		if !ok {
			matIdx = len(colorToMat)
			colorToMat[color32] = matIdx
			matName := "mat" + strconv.Itoa(matIdx)
			m.Materials = append(m.Materials, &fileformats.MTLFileMaterial{
				Name:    matName,
				Ambient: color32,
				Diffuse: color32,
			})
			group = &fileformats.OBJFileFaceGroup{Material: matName}
			o.FaceGroups = append(o.FaceGroups, group)
		} else {
			group = o.FaceGroups[matIdx]
		}
		face := [3][3]int{}
		for i, p := range tri {
			if idx, ok := coordToIdx.Load(p); !ok {
				idx = coordToIdx.Len()
				coordToIdx.Store(p, idx)
				o.Vertices = append(o.Vertices, p.Array())
				face[i][0] = idx + 1
			} else {
				face[i][0] = idx + 1
			}
		}
		group.Faces = append(group.Faces, face)
	}

	return
}

// VertexColorsToTriangle creates a per-triangle color
// function that averages the colors at each of the
// vertices.
func VertexColorsToTriangle(f func(c Coord3D) [3]float64) func(t *Triangle) [3]float64 {
	return func(t *Triangle) [3]float64 {
		var sum [3]float64
		for _, c := range t {
			color := f(c)
			for i, x := range color {
				sum[i] += x / 3
			}
		}
		return sum
	}
}
