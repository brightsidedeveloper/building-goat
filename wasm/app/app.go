package app

import (
	"context"
	"fmt"
	"wasm/goat"
)

func App(ctx context.Context, props any) goat.VNode {
	return goat.VNode{
		Key:      "lol",
		Tag:      "div",
		Attrs:    map[string]string{"style": "color: red;"},
		Children: []goat.VNode{{Text: fmt.Sprintf("Hello, World!")}},
	}
}

// func App(ctx context.Context, props any) goat.VNode {
// 	count, setCount := goat.UseState(ctx, 0)
// 	goat.UseEffect(ctx, func() func() {
// 		goat.Log("Count updated to ", count())
// 		return func() {}
// 	}, count())
// 	return goat.VNode{
// 		Tag: "div",
// 		Attrs: map[string]string{
// 			"style": "color:red;",
// 		},
// 		Children: []goat.VNode{
// 			{Text: fmt.Sprintf("Count: %d", count())},
// 			{Tag: "button",
// 				Events: map[string]js.Func{
// 					"click": js.FuncOf(func(this js.Value, args []js.Value) any {
// 						setCount(count() + 1)
// 						return nil
// 					})},
// 				Children: []goat.VNode{
// 					{Text: "Click Me"},
// 				},
// 			},
// 		},
// 	}
// }
