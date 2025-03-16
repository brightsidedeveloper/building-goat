package main

import (
	"wasm/app/components"

	"github.com/brightsidedeveloper/goat"
)

func main() {
	done := make(chan struct{})
	props := components.AppProps{
		Counters: []components.CounterProps{
			{InitialCount: 5},
			{InitialCount: 10},
		},
	}
	goat.RenderRoot("root", components.App, props)
	<-done
}
