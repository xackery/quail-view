package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/xackery/quail-view/anim"
	"github.com/xackery/quail-view/skeleton"

	"github.com/xackery/quail-view/mesh"

	"github.com/xackery/quail/quail"

	"github.com/xackery/engine/app"
	"github.com/xackery/engine/camera"
	"github.com/xackery/engine/core"
	"github.com/xackery/engine/gls"
	"github.com/xackery/engine/graphic"
	"github.com/xackery/engine/gui"
	"github.com/xackery/engine/gui/assets/icon"
	"github.com/xackery/engine/light"
	"github.com/xackery/engine/loader/collada"
	"github.com/xackery/engine/loader/obj"
	"github.com/xackery/engine/math32"
	"github.com/xackery/engine/renderer"
	"github.com/xackery/engine/util/helper"
	"github.com/xackery/engine/window"
)

var (
	// Version is the application version
	Version = "0.0.0"
	gv      *g3nView
)

const (
	checkON  = icon.CheckBox
	checkOFF = icon.CheckBoxOutlineBlank
)

type g3nView struct {
	*app.Application                // Embedded application object
	fs               *FileSelect    // File selection dialog
	ed               *ErrorDialog   // Error dialog
	axes             *helper.Axes   // Axis helper
	grid             *helper.Grid   // Grid helper
	viewAxes         bool           // Axis helper visible flag
	viewGrid         bool           // Grid helper visible flag
	camPos           math32.Vector3 // Initial camera position
	models           []*core.Node   // Models being shown
	scene            *core.Node
	cam              *camera.Camera
	fpsCam           *camera.Camera
	focusMenu        *gui.Menu
	focusModels      []*focusEntry
	orbit            *camera.OrbitControl
}

type focusEntry struct {
	node      *core.Node
	meshWidth float64
	mi        *gui.MenuItem
}

func (e *focusEntry) onClick(evname string, ev interface{}) {
	if e.node == nil {
		fmt.Println("Node not found for focus entry")
		return
	}

	pos := e.node.Position()
	//gv.cam.LookAt(&pos, &math32.Vector3{X: 0, Y: 1, Z: 0})
	pos.Z += float32(e.meshWidth / 2)
	// pos.Z += 10
	// gv.cam.SetPositionVec(&pos)
	// rot := e.node.Rotation()
	// gv.cam.SetRotationVec(&rot)
	gv.orbit.SetTarget(pos)
	gv.cam.LookAt(&pos, &math32.Vector3{X: 0, Y: 1, Z: 0})
	gv.cam.SetPositionVec(&pos)
	gv.orbit.Reset()

	fmt.Println("Focusing on", e.node.Name())
}

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

	gv = &g3nView{}

	// Create application and scene
	a := app.App(600, 600, fmt.Sprintf("quail-view v%s - %s", Version, filepath.Base(path)))
	gv.Application = a

	scene := core.NewNode()
	gv.scene = scene

	gv.axes = helper.NewAxes(2)
	gv.viewAxes = true
	gv.axes.SetVisible(gv.viewAxes)
	gv.scene.Add(gv.axes)

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Adds a grid helper to the scene initially not visible
	gv.grid = helper.NewGrid(50, 1, &math32.Color{R: 0.4, G: 0.4, B: 0.4})
	gv.viewGrid = true
	gv.grid.SetVisible(gv.viewGrid)
	gv.scene.Add(gv.grid)

	gv.camPos = math32.Vector3{X: 8.3, Y: 4.7, Z: 3.7}
	gv.cam = camera.New(1)
	gv.cam.SetPositionVec(&gv.camPos)
	gv.cam.LookAt(&math32.Vector3{X: 0, Y: 0, Z: 0}, &math32.Vector3{X: 0, Y: 1, Z: 0})
	gv.orbit = camera.NewOrbitControl(gv.cam)

	// Create perspective camera
	gv.fpsCam = camera.New(1)
	gv.fpsCam.SetFar(50000)
	gv.fpsCam.SetNear(0.1)
	gv.fpsCam.SetProjection(camera.Perspective)
	gv.fpsCam.SetVisible(false)
	scene.Add(gv.fpsCam)

	// Set up orbit control for the camera
	camera.NewFlyControl(gv.fpsCam, &math32.Vector3{X: 0, Y: -10, Z: 0}, &math32.Vector3{X: 0, Y: 0, Z: 1}, camera.FlightSimStyle())

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		gv.cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	gv.buildGui()

	var err error

	q := &quail.Quail{}
	err = q.PfsRead(path)
	if err != nil {
		return fmt.Errorf("pfs read: %w", err)
	}

	maxWidth := 3.0
	riggedMeshes := make([]*graphic.RiggedMesh, 0)

	for i := 0; i < len(q.Models); i++ {
		var meshInstance core.INode
		model := q.Models[i]
		mesh, err := mesh.Generate(q, model)
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

		if len(gv.focusModels) > i {
			fm := gv.focusModels[i]
			fm.meshWidth = meshWidth
			fm.node = node
			fm.mi.SetVisible(true)
			fm.mi.SetText(model.Header.Name)
		}
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
	gv.cam.SetPosition(0, 0, float32(maxWidth))

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
	//scene.Add(helper.NewAxes(0.5))

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
		renderer.Render(scene, gv.cam)
		if len(anims) > 0 {
			anims[0].Update(float32(deltaTime.Seconds()))
		}
	})
	return nil
}

// setupGui builds the GUI
func (gv *g3nView) buildGui() error {

	gui.Manager().Set(gv.scene)

	// Adds menu bar
	mb := gui.NewMenuBar()
	mb.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 0, AlignH: gui.AlignWidth})
	gv.scene.Add(mb)

	// Create "File" menu and adds it to the menu bar
	m1 := gui.NewMenu()
	m1.AddOption("Open model").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.fs.Show(true)
	})
	m1.AddOption("Remove models").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.removeModels()
	})
	m1.AddOption("Reset camera").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.cam.SetPositionVec(&gv.camPos)
		gv.cam.LookAt(&math32.Vector3{X: 0, Y: 0, Z: 0}, &math32.Vector3{X: 0, Y: 1, Z: 0})
		gv.orbit.Reset()
	})
	m1.AddSeparator()
	m1.AddOption("Quit").SetId("quit").Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.Exit()
	})
	mb.AddMenu("File", m1)

	// Create "View" menu and adds it to the menu bar
	m2 := gui.NewMenu()
	vAxis := m2.AddOption("View axis helper").SetIcon(checkOFF)
	vAxis.SetIcon(getIcon(gv.viewAxes))
	vAxis.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.viewAxes = !gv.viewAxes
		vAxis.SetIcon(getIcon(gv.viewAxes))
		gv.axes.SetVisible(gv.viewAxes)
	})

	vGrid := m2.AddOption("View grid helper").SetIcon(checkOFF)
	vGrid.SetIcon(getIcon(gv.viewGrid))
	vGrid.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		gv.viewGrid = !gv.viewGrid
		vGrid.SetIcon(getIcon(gv.viewGrid))
		gv.grid.SetVisible(gv.viewGrid)
	})

	mb.AddMenu("View", m2)

	gv.focusMenu = gui.NewMenu()
	for i := 0; i < 10; i++ {
		fe := &focusEntry{}

		fe.mi = gv.focusMenu.AddOption(fmt.Sprintf("Focus %d", i))
		fe.mi.Subscribe(gui.OnClick, fe.onClick)
		fe.mi.SetVisible(false)
		gv.focusModels = append(gv.focusModels, fe)
	}
	mb.AddMenu("Center On", gv.focusMenu)
	// vView := m2.AddOption("Toggle View Mode").SetIcon(checkOFF)
	// vView.SetIcon(getIcon(gv.isFpsCamera))
	// vView.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
	// 	gv.isFpsCamera = !gv.isFpsCamera
	// 	vView.SetIcon(getIcon(gv.isFpsCamera))
	// 	if gv.isFpsCamera {
	// 		pos := gv.cam.Position()
	// 		gv.fpsCam.SetPositionVec(&pos)
	// 		rot := gv.cam.Rotation()
	// 		gv.fpsCam.SetRotationVec(&rot)
	// 	} else {
	// 		pos := gv.fpsCam.Position()
	// 		gv.cam.SetPositionVec(&pos)
	// 		rot := gv.fpsCam.Rotation()
	// 		gv.cam.SetRotationVec(&rot)
	// 	}
	// 	gv.cam.SetVisible(!gv.isFpsCamera)
	// 	gv.fpsCam.SetVisible(gv.isFpsCamera)
	// })

	// Creates file selection dialog
	fs, err := NewFileSelect(400, 300)
	if err != nil {
		return err
	}
	gv.fs = fs
	gv.fs.SetVisible(false)
	gv.fs.Subscribe("OnOK", func(evname string, ev interface{}) {
		fpath := gv.fs.Selected()
		if fpath == "" {
			gv.ed.Show("File not selected")
			return
		}
		err := gv.openModel(fpath)
		if err != nil {
			gv.ed.Show(err.Error())
			return
		}
		gv.fs.SetVisible(false)

	})
	gv.fs.Subscribe("OnCancel", func(evname string, ev interface{}) {
		gv.fs.Show(false)
	})
	gv.scene.Add(gv.fs)

	// Creates error dialog
	gv.ed = NewErrorDialog(600, 100)
	gv.scene.Add(gv.ed)

	return nil
}

// openModel try to open the specified model and add it to the scene
func (gv *g3nView) openModel(fpath string) error {

	dir, file := filepath.Split(fpath)
	ext := filepath.Ext(file)

	// Loads OBJ model
	if ext == ".obj" {
		// Checks for material file in the same dir
		matfile := file[:len(file)-len(ext)]
		matpath := filepath.Join(dir, matfile)
		_, err := os.Stat(matpath)
		if err != nil {
			matpath = ""
		}

		// Decodes model in in OBJ format
		dec, err := obj.Decode(fpath, matpath)
		if err != nil {
			return err
		}

		// Creates a new node with all the objects in the decoded file and adds it to the scene
		group, err := dec.NewGroup()
		if err != nil {
			return err
		}
		gv.scene.Add(group)
		gv.models = append(gv.models, group)
		return nil
	}

	// Loads COLLADA model
	if ext == ".dae" {
		dec, err := collada.Decode(fpath)
		if err != nil && err != io.EOF {
			return err
		}
		dec.SetDirImages(dir)

		// Loads collada scene
		s, err := dec.NewScene()
		if err != nil {
			return err
		}
		gv.scene.Add(s)
		gv.models = append(gv.models, s.GetNode())
		return nil
	}
	return fmt.Errorf("Unrecognized model file extension:[%s]", ext)
}

// removeModels removes and disposes of all loaded models in the scene
func (gv *g3nView) removeModels() {
	for i := 0; i < len(gv.models); i++ {
		model := gv.models[i]
		gv.scene.Remove(model)
		model.Dispose()
	}
	gv.models = nil
}

func getIcon(state bool) string {

	if state {
		return icon.CheckBox
	} else {
		return icon.CheckBoxOutlineBlank
	}
}
