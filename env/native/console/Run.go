package console

import (
	"log"
	"runtime"

	"github.com/jroimartin/gocui"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/native"
)

// Run initializes the environment to run the given application within.
func Run(app env.Application, deferrer <-chan func()) {
	runtime.LockOSThread()

	gui := gocui.NewGui()
	if err := gui.Init(); err != nil {
		log.Panicln(err)
	}
	defer gui.Close()

	gui.SetLayout(layout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	var window *native.OpenGlWindow
	{
		var err error
		window, err = native.NewOpenGlWindow()
		if err != nil {
			log.Panicln(err)
		}
	}

	app.Init(window)

	startDeferrerRoutine(gui, deferrer)

	gui.Execute(getWindowUpdater(window))
	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	window.Close()
}

func getWindowUpdater(window *native.OpenGlWindow) (updater func(*gocui.Gui) error) {
	updater = func(gui *gocui.Gui) error {
		window.Update()
		gui.Execute(updater)

		return nil
	}

	return
}

func startDeferrerRoutine(gui *gocui.Gui, deferrer <-chan func()) {
	go func() {
		for task := range deferrer {
			deferredTask := task
			gui.Execute(func(*gocui.Gui) error {
				deferredTask()

				return nil
			})
		}
	}()
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("side", -1, -1, int(0.2*float32(maxX)), maxY-5); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	if _, err := g.SetView("main", int(0.2*float32(maxX)), -1, maxX, maxY-5); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	if _, err := g.SetView("cmdline", -1, maxY-5, maxX, maxY); err != nil &&
		err != gocui.ErrUnknownView {
		return err
	}
	return nil
}
