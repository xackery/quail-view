package game

/*
import (
	"fmt"
	"g3nd/app"
	"path/filepath"

	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/window"
)

type Game struct {
	app   *app.App
	scene *core.Node
}

func New(path string, Version string) *Game {
	g := &Game{
		app:   app.App(600, 600, fmt.Sprintf("quail-view v%s - %s", Version, filepath.Base(path))),
		scene: core.NewNode(),
	}

	width, height := g.GetSize()
	aspect := float32(width) / float32(height)
	a.camera = camera.New(aspect)

	g.Subscribe(window.OnWindowSize, func(evname string, ev interface{}) { a.OnWindowResize() })
	g.OnWindowResize()
	// Subscribe to key events
	a.Subscribe(window.OnKeyDown, func(evname string, ev interface{}) {
		kev := ev.(*window.KeyEvent)
		if kev.Key == window.KeyEscape { // ESC terminates the program
			a.Exit()
		} else if kev.Key == window.KeyF11 { // F11 toggles full screen
			//a.Window().SetFullScreen(!a.Window().FullScreen()) // TODO
		} else if kev.Key == window.KeyS && kev.Mods == window.ModAlt { // Ctr-S prints statistics in the console
			a.logStats()
		}
	})

	return g
}

// OnWindowResize is default handler for window resize events.
func (g *Game) OnWindowResize() {

	// Get framebuffer size and set the viewport accordingly
	width, height := g.GetFramebufferSize()
	g.app.Gls().Viewport(0, 0, int32(width), int32(height))

	// Set camera aspect ratio
	g.camera.SetAspect(float32(width) / float32(height))

}

func (a *Game) Run() {

	g.app.Run(g.Update)
}
*/
