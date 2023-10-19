package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xackery/quail/pfs"

	"github.com/xackery/quail-view/mesh"
	"github.com/xackery/quail/quail"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

var (
	// Version is the application version
	Version = "dev"
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

	path := os.Args[1]

	// Create application and scene
	a := app.App()

	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 3)
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

	// Create a blue torus and add it to the scene
	//geom := geometry.NewPlane(1, 1)
	//geom := plane.NewPlane(1, 1)
	for i := 0; i < len(q.Models); i++ {
		mesh, err := mesh.Generate(archive, q.Models[i])
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		scene.Add(mesh)
	}

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{R: 1.0, G: 1.0, B: 1.0}, 0.8))
	pointLight := light.NewPoint(&math32.Color{R: 1, G: 1, B: 1}, 5.0)
	pointLight.SetPosition(1, 0, 2)
	scene.Add(pointLight)

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(0.5))

	// Set background color to black
	a.Gls().ClearColor(0, 0, 0, 1)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
	return nil
}
