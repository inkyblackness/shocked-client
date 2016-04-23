package console

import (
	"fmt"
	"log"
	"runtime"

	"github.com/jroimartin/gocui"

	"github.com/inkyblackness/shocked-client/env"
	"github.com/inkyblackness/shocked-client/env/native"
	//"github.com/inkyblackness/shocked-client/viewmodel"
)

type appRunner struct {
	app env.Application
}

// Run initializes the environment to run the given application within.
func Run(app env.Application, deferrer <-chan func()) {
	runtime.LockOSThread()

	gui := gocui.NewGui()
	if err := gui.Init(); err != nil {
		log.Panicln(err)
	}
	defer gui.Close()

	runner := &appRunner{app: app}

	gui.Cursor = true
	gui.Mouse = true
	gui.SelBgColor = gocui.ColorGreen
	gui.SelFgColor = gocui.ColorBlack
	gui.SetLayout(runner.layout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := gui.SetKeybinding("mainSection", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := gui.SetKeybinding("mainSection", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
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

func (runner *appRunner) layout(g *gocui.Gui) error {
	//vm := runner.app.ViewModel()
	maxX, maxY := g.Size()

	if view, err := g.SetView("mainSection", -1, -1, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Highlight = true
		/*
			selectedMainSection := vm.(*viewmodel.ContainerNode).Get()["view"].(*viewmodel.ContainerNode).Get()["selectedMainSection"].(*viewmodel.StringValueNode).Get()
			fmt.Fprintf(view, "v Main Section: <%v>\n", selectedMainSection)
		*/
		fmt.Fprintf(view, "Line1\n")
		fmt.Fprintf(view, "Extra Line\n")
		g.SetCurrentView("mainSection")
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}
