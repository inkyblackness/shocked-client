package main

import (
	"github.com/inkyblackness/shocked-client/editor"
	"github.com/inkyblackness/shocked-client/env/native/console"
)

func main() {
	app := editor.NewMainApplication(nil)

	console.Run(app)
}
