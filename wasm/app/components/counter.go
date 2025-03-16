// components/counter.go
package components

import (
	"context"
	"strconv"
	"syscall/js"

	"github.com/brightsidedeveloper/goat"
)

type CounterProps struct {
	InitialCount int
}

func Counter(ctx context.Context, props any) goat.VNode {
	counterProps := props.(CounterProps)
	count, setCount := goat.UseState[int](ctx, counterProps.InitialCount)
	increment := goat.UseCallback(ctx, func(this js.Value, args []js.Value) any {
		setCount(count() + 1)
		return nil
	})
	goat.UseEffect(ctx, func() {
		js.Global().Get("console").Call("log", "Counter rendered with count:", count())
	})

	return goat.NewVNode("div", nil, nil, []goat.VNode{
		goat.NewVNode("", nil, nil, nil, strconv.Itoa(count())),
		goat.NewVNode("button", nil, map[string]func(js.Value, []js.Value) any{
			"onclick": increment,
		}, nil, "Increment"),
	}, "")
}
