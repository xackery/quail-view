package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xackery/quail-view/anim"
	"github.com/xackery/quail-view/skeleton"

	"github.com/xackery/quail-view/mesh"

	"github.com/xackery/quail/pfs"

	"github.com/xackery/quail/quail"

	"github.com/xackery/engine/app"
	"github.com/xackery/engine/camera"
	"github.com/xackery/engine/core"
	"github.com/xackery/engine/gls"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/engine/gui"
	"github.com/xackery/engine/light"
	"github.com/xackery/engine/math32"
	"github.com/xackery/engine/renderer"
	"github.com/xackery/engine/util/helper"
	"github.com/xackery/engine/window"
)

var (
	// Version is the application version
	Version = "0.0.0"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println("Failed to run:", err)
		os.Exit(1)
	}
}

func run() error {

	if len(os.Args) < 2 {
		return fmt.Errorf("no file specified")
	}

	quail.SetLogLevel(2)
	path := os.Args[1]

	// Create application and scene
	a := app.App(600, 600, fmt.Sprintf("quail-view v%s - %s", Version, filepath.Base(path)))

	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetFar(50000)
	cam.SetNear(0.1)
	cam.SetProjection(camera.Perspective)
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	var err error
	q := quail.New()
	switch filepath.Ext(path) {
	case ".s3d":
		err = q.S3DImport(path)
	case ".eqg":
		err = q.EQGImport(path)
	}

	if err != nil {
		return fmt.Errorf("eqg import: %w", err)
	}
	archive, err := pfs.NewFile(path)
	if err != nil {
		return fmt.Errorf("eqg load: %w", err)
	}
	defer archive.Close()

	maxWidth := 3.0
	riggedMeshes := make([]*graphic.RiggedMesh, 0)
	for i := 0; i < len(q.Models); i++ {
		var meshInstance core.INode
		model := q.Models[i]
		mesh, err := mesh.Generate(archive, model)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		mesh.SetPosition(0, 0, float32(float64(i)*2.0))

		meshInstance = mesh

		if len(model.Bones) > 0 {
			skel, err := skeleton.Generate(q.Models[i].Bones)
			if err != nil {
				return fmt.Errorf("generate skeleton: %w", err)
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

		node := scene.Add(meshInstance)
		node.SetName(model.Header.Name)
	}

	fmt.Println("total rigged meshes:", len(riggedMeshes))
	anims, err := anim.Generate(q.Animations, riggedMeshes)
	if err != nil {
		return fmt.Errorf("generate anim: %w", err)
	}

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{R: 1.0, G: 1.0, B: 1.0}, 2)) //0.8

	base := float32(5.0)

	pointLight := light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, base*float32(maxWidth)*0.5)
	pointLight.SetPosition(1, 0, float32(maxWidth/2))
	scene.Add(pointLight)
	cam.SetPosition(0, 0, float32(maxWidth))

	pointLight = light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, base*float32(maxWidth)*0.5)
	pointLight.SetPosition(float32(maxWidth/2), 0, 0)
	scene.Add(pointLight)

	pointLight = light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, base*float32(maxWidth)*0.5)
	pointLight.SetPosition(0, float32(maxWidth/2), 0)
	scene.Add(pointLight)

	dir1 := light.NewDirectional(&math32.Color{R: 1, G: 1, B: 1}, 1.0)
	dir1.SetPosition(0, 5, 10)
	scene.Add(dir1)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	a.Gls().ClearColor(0.2, 0.2, 0.2, 1)

	/*
		panel := gui.NewPanel(150, 30)
		panel.SetColor4(&gui.StyleDefault().Scroller.BgColor)
		panel.SetLayoutParams(&gui.DockLayoutParams{Edge: gui.DockCenter})
		panel.SetRenderable(true)
		panel.SetEnabled(true)
		scene.Add(panel)
		//gui.Manager().Set(panel)

		mbOption := gui.NewLabel("Layer: ")
		mbOption.SetPosition(80, 10)
		mbOption.SetPaddings(2, 2, 2, 2)
		mbOption.SetBorders(1, 1, 1, 1)
		panel.Add(mbOption)

		mb := gui.NewMenuBar()
		mb.SetPosition(10, 10)
		layerMenu := gui.NewMenu()
		layerMenu.AddOption("Layer 0").SetId("layer0")
		layerMenu.AddOption("Layer 1").SetId("layer1")
		layerMenu.AddOption("Layer 2").SetId("layer2")
		mb.AddMenu("Layer", layerMenu)
		panel.Add(mb)
	*/

	//a.IWindow.(*window.GlfwWindow).SetTitle(fmt.Sprintf("quail-view v%s - %s", Version, filepath.Base(path)))

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
		if len(anims) > 0 {
			anims[0].Update(float32(deltaTime.Seconds()))
		}
	})
	return nil
}
