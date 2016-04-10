package main

import (
	"github.com/inkyblackness/shocked-client/editor"
	"github.com/inkyblackness/shocked-client/env/native"
	"github.com/inkyblackness/shocked-client/env/native/console"
)

func main() {
	deferrer := make(chan func(), 100)
	defer close(deferrer)

	transport := native.NewRestTransport("http://localhost:8080", deferrer)
	store := editor.NewRestDataStore(transport)
	app := editor.NewMainApplication(store)

	console.Run(app, deferrer)
}
