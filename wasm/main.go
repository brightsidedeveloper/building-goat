package main

import (
	"wasm/app"
	"wasm/goat"
)

func main() {
	goat.RenderRoot("root", app.App, nil)
}
