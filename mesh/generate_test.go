package mesh

import (
	"fmt"
	"os"
	"testing"

	"github.com/xackery/engine/core"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/quail-view/skeleton"
	"github.com/xackery/quail/quail"
)

func TestMesh(t *testing.T) {
	q := &quail.Quail{}
	path := os.Getenv("EQ_PATH")
	if path == "" {
		t.Fatalf("EQ_PATH not set")
	}
	err := q.PfsRead(path + "/ael_chr.s3d")
	if err != nil {
		t.Fatalf("pfs read: %s", err.Error())
	}

	maxWidth := 3.0
	riggedMeshes := make([]*graphic.RiggedMesh, 0)

	for i := 0; i < len(q.Models); i++ {
		var meshInstance core.INode
		model := q.Models[i]
		mesh, err := Generate(q, model)
		if err != nil {
			t.Fatalf("generate: %s", err.Error())
		}

		mesh.SetPosition(0, 0, float32(float64(i)*2.0))

		meshInstance = mesh

		if len(model.Bones) > 0 {
			skel, err := skeleton.Generate(q.Models[i].Bones)
			if err != nil {
				t.Fatalf("generate skeleton: %s", err.Error())
			}

			rigMesh := graphic.NewRiggedMesh(mesh)
			rigMesh.SetSkeleton(skel)
			meshInstance = rigMesh
			riggedMeshes = append(riggedMeshes, rigMesh)
		}

		meshWidth := float64(mesh.BoundingBox().Max.X) * 2
		if float64(mesh.BoundingBox().Max.Y)*2 > meshWidth {
			meshWidth = float64(mesh.BoundingBox().Max.Y) * 2
		}
		if float64(mesh.BoundingBox().Max.Z)*2 > meshWidth {
			meshWidth = float64(mesh.BoundingBox().Max.Z) * 2
		}

		if meshWidth > maxWidth {
			maxWidth = meshWidth
		}

		fmt.Println("done", meshInstance)
	}
}
