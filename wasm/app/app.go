package app

import (
	"context"
	"fmt"
	"syscall/js"
	"wasm/goat"
)

type CounterProps struct {
	InitialCount int
}

func Counter(ctx context.Context, props any) goat.VNode {
	p := props.(CounterProps)
	count, setCount := goat.UseState(ctx, p.InitialCount)
	goat.UseEffect(ctx, func() func() {
		js.Global().Get("console").Call("log", "Counter mounted")
		return func() {
			js.Global().Get("console").Call("log", "Counter unmounted")
		}
	}, count())
	increment := js.FuncOf(func(this js.Value, args []js.Value) any {
		setCount(count() + 1)
		return nil
	})
	return goat.VNode{
		Tag: "div",
		Children: []goat.VNode{
			{Text: "Count: " + fmt.Sprint(count())},
			{
				Tag:    "button",
				Events: map[string]js.Func{"click": increment},
				Text:   "Increment",
			},
		},
		Key: fmt.Sprint(p.InitialCount),
	}
}

func App(ctx context.Context, props any) goat.VNode {
	return goat.VNode{
		Tag:   "div",
		Attrs: map[string]string{"style": "display: flex; flex-direction: column;"},
		Children: []goat.VNode{
			{Component: Counter, Props: CounterProps{InitialCount: 0}, Key: "c0"},
			{Component: Counter, Props: CounterProps{InitialCount: 1}, Key: "c1"},
		},
	}
}
