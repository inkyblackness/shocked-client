package main

import (
	"github.com/inkyblackness/shocked-client/editor"
	"github.com/inkyblackness/shocked-client/env/browser"
)

func main() {
	app := editor.NewMainApplication()

	browser.Run(app)
}
