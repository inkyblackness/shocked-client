package native

import (
	"log"
	"runtime"

	"github.com/inkyblackness/shocked-client/env"
)

// Run initializes the environment to run the given application within.
func Run(app env.Application, deferrer <-chan func()) {
	runtime.LockOSThread()

	var window *OpenGlWindow
	{
		var err error
		window, err = NewOpenGlWindow()
		if err != nil {
			log.Panicln(err)
		}
	}

	app.Init(window)

	for !window.ShouldClose() {
		select {
		case task := <-deferrer:
			{
				task()
			}
		default:
		}
		window.Update()
	}

	window.Close()
}
