package mesh

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/malashin/dds"
	"github.com/sergeymakinen/go-bmp"
	"github.com/xackery/colors"
	"github.com/xackery/engine/texture"

	"github.com/xackery/quail/common"

	"github.com/xackery/engine/geometry"
	"github.com/xackery/engine/gls"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/engine/material"
	"github.com/xackery/engine/math32"
)

var (
	fallbackImg *image.RGBA
)

func Generate(in *common.Model) (*graphic.Mesh, error) {
	mats := make(map[string]*material.Standard)
	matIndexes := make(map[string]int)

	for _, mat := range in.Materials {
		newMat, ok := mats[mat.Name]
		if !ok {
			newMat = material.NewStandard(math32.NewColor("gray"))
			//newMat.SetShader("MaxCB1")
			//newMat.SetShader("MPLBasic")
			matIndexes[mat.Name] = len(mats)
			mats[mat.Name] = newMat
		}

		for _, property := range mat.Properties {
			if property.Category != 2 {
				continue
			}

			if !strings.Contains(strings.ToLower(property.Name), "texture") {
				continue
			}

			img, err := generateImage(property.Value, property.Data)
			if err != nil {
				return nil, fmt.Errorf("generate image: %w", err)
			}

			newMat.AddTexture(texture.NewTexture2DFromRGBA(img))
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

	//fmt.Printf("%d total materials, %d triangles\n", len(matIndexes), len(in.Triangles))

	return mesh, nil
}

func generateImage(name string, data []byte) (*image.RGBA, error) {
	if len(data) == 0 {
		fmt.Println("empty texture", name, "fallback pink image")
		return fallback(), nil
	}

	if string(data[0:3]) == "DDS" {
		// change to png, blender doesn't like EQ dds
		img, err := dds.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Println("Failed to decode dds:", name, err, "fallback pink image")
			return fallback(), nil
		}
		switch rgba := img.(type) {
		case *image.RGBA:
			return rgba, nil
		case *image.NRGBA:
			newImg := image.NewRGBA(rgba.Rect)
			draw.Draw(newImg, newImg.Bounds(), rgba, rgba.Rect.Min, draw.Src)
			return newImg, nil
		default:
			return nil, fmt.Errorf("unknown dds type %T", rgba)
		}
	}

	if filepath.Ext(strings.ToLower(name)) == ".png" {
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			fmt.Println("Failed to decode png:", name, err, "fallback pink image")
			return fallback(), nil
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
			fmt.Println("Failed to decode bmp:", name, err, "fallback pink image")
			return fallback(), nil
		}
		switch rgba := img.(type) {
		case *image.RGBA:
			return rgba, nil
		case *image.NRGBA:
			return image.NewRGBA(rgba.Rect), nil
		default:
			fmt.Println("Failed dds type", rgba, "fallback pink image")
			return fallback(), nil
		}
	}
	return nil, fmt.Errorf("unknown image type %s", name)
}

func fallback() *image.RGBA {
	if fallbackImg != nil {
		return fallbackImg
	}
	var img image.Image
	img = image.NewRGBA(image.Rect(0, 0, 64, 64))
	dimg := img.(draw.Image)

	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			dimg.Set(x, y, colors.Magenta)
		}
	}
	fallbackImg = img.(*image.RGBA)
	return fallbackImg
}
