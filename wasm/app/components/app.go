// components/app.go
package components

import (
	"context"

	"github.com/brightsidedeveloper/goat"
)

type AppProps struct {
	Counters []CounterProps
}

func App(ctx context.Context, props any) goat.VNode {
	appProps := props.(AppProps)
	children := make([]goat.VNode, len(appProps.Counters))
	for i, counterProps := range appProps.Counters {
		children[i] = Counter(ctx, counterProps)
	}
	return goat.NewVNode("div", nil, nil, children, "")
}
