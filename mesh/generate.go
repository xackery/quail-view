package mesh

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/malashin/dds"
	"github.com/sergeymakinen/go-bmp"
	"github.com/xackery/quail/pfs"

	"github.com/xackery/quail/common"

	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
)

func Generate(archive *pfs.PFS, in *common.Model) (*graphic.Mesh, error) {
	mats := make(map[string]*material.Standard)
	matIndexes := make(map[string]int)

	for _, mat := range in.Materials {
		for _, property := range mat.Properties {
			if property.Category != 2 {
				continue
			}
			if !strings.Contains(strings.ToLower(property.Name), "texture") {
				continue
			}
			data, err := archive.File(property.Value)
			if err != nil {
				continue
				//return nil, fmt.Errorf("file %s: %w", property.Value, err)
			}
			img, err := generateImage(property.Value, data)
			if err != nil {
				return nil, fmt.Errorf("generate image: %w", err)
			}

			newMat, ok := mats[mat.Name]
			if !ok {
				newMat = material.NewStandard(math32.NewColor("gray"))
				matIndexes[mat.Name] = len(mats)
				mats[mat.Name] = newMat
			}

			if img != nil {

			}
			//newMat.AddTexture(texture.NewTexture2DFromRGBA(img))
		}

	}

	geom := geometry.NewGeometry()

	positions := math32.NewArrayF32(0, 16)
	normals := math32.NewArrayF32(0, 16)
	uvs := math32.NewArrayF32(0, 16)
	indices := math32.NewArrayU32(0, 16)

	for i := 0; i < len(in.Vertices); i++ {
		positions.Append(float32(in.Vertices[i].Position.X), float32(in.Vertices[i].Position.Y), float32(in.Vertices[i].Position.Z))
		normals.Append(float32(in.Vertices[i].Normal.X), float32(in.Vertices[i].Normal.Y), float32(in.Vertices[i].Normal.Z))
		uvs.Append(float32(in.Vertices[i].Uv.X), float32(in.Vertices[i].Uv.Y))
	}

	lastMat := ""
	lastIndex := 0
	for i := 0; i < len(in.Triangles); i++ {
		indices.Append(uint32(in.Triangles[i].Index.X), uint32(in.Triangles[i].Index.Y), uint32(in.Triangles[i].Index.Z))
		geom.AddGroup(lastIndex, i-lastIndex, matIndexes[lastMat])
		lastMat = in.Triangles[i].MaterialName
		lastIndex = i
	}

	geom.SetIndices(indices)
	geom.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	geom.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	geom.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))

	//mat := material.NewStandard(math32.NewColor("DarkBlue"))
	mesh := graphic.NewMesh(geom, nil)

	for name, idx := range matIndexes {
		mesh.AddGroupMaterial(mats[name], idx)
	}

	fmt.Printf("%d total materials, %d triangles\n", len(matIndexes), len(in.Triangles))

	return mesh, nil
}

func generateImage(name string, data []byte) (*image.RGBA, error) {
	if string(data[0:3]) == "DDS" {
		// change to png, blender doesn't like EQ dds
		img, err := dds.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("dds decode: %w", err)
		}
		switch rgba := img.(type) {
		case *image.RGBA:
			return rgba, nil
		case *image.NRGBA:
			return image.NewRGBA(rgba.Rect), nil
		default:
			return nil, fmt.Errorf("unknown dds type %T", rgba)
		}
	}

	if filepath.Ext(strings.ToLower(name)) == ".png" {
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("png decode: %w", err)
		}
		switch rgba := img.(type) {
		case *image.RGBA:
			return rgba, nil
		case *image.NRGBA:
			return image.NewRGBA(rgba.Rect), nil
		default:
			return nil, fmt.Errorf("unknown dds type %T", rgba)
		}
	}

	if filepath.Ext(strings.ToLower(name)) == ".bmp" {

		img, err := bmp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("bmp decode: %w", err)
		}
		switch rgba := img.(type) {
		case *image.RGBA:
			return rgba, nil
		case *image.NRGBA:
			return image.NewRGBA(rgba.Rect), nil
		default:
			return nil, fmt.Errorf("unknown dds type %T", rgba)
		}
	}
	return nil, fmt.Errorf("unknown image type %s", name)
}
