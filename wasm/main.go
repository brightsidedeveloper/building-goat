package main

import (
	"wasm/app/components"
	"wasm/goat"
)

func main() {

	goat.RenderRoot("root", components.App, nil)

	done := make(chan struct{})
	<-done
}
