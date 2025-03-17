package goat

import (
	"context"
	"sync"
	"syscall/js"
)

type Fiber struct {
	Component  Component
	Props      any
	State      map[int]any
	StateOrder []int
	CallIndex  int
	Effects    []Effect
	Parent     *Fiber
	Child      *Fiber
	Sibling    *Fiber
	Alternate  *Fiber
	Key        string
	Node       VNode
	DOM        js.Value
	Mu         sync.Mutex
	Dirty      bool
	Renderer   *Renderer
}

type Component func(ctx context.Context, props any) VNode

type Effect struct {
	Setup   func() func()
	Cleanup func()
	Deps    []any
}

func NewFiber(comp Component, props any, parent *Fiber, key string) *Fiber {
	return &Fiber{
		Component:  comp,
		Props:      props,
		State:      make(map[int]any),
		StateOrder: []int{},
		Effects:    []Effect{},
		Parent:     parent,
		Key:        key,
		Dirty:      true,
	}
}

// Ask how this works!
var fiberKey = struct{}{}

func GetFiberFromContext(ctx context.Context) *Fiber {
	if f, ok := ctx.Value(fiberKey).(*Fiber); ok {
		return f
	}
	panic("No fiber found in context")
}

func (r *Renderer) RenderFiber(fiber *Fiber) {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	if !fiber.Dirty {
		return
	}

	if fiber.Alternate == nil {
		fiber.Alternate = NewFiber(fiber.Component, fiber.Props, fiber.Parent, fiber.Key)
	}

	fiber.Mu.Lock()
	fiber.CallIndex = 0
	ctx := context.WithValue(context.Background(), fiberKey, fiber)
	fiber.Node = fiber.Component(ctx, fiber.Props)
	fiber.Dirty = false
	fiber.Mu.Unlock()

	r.ReconcileChildren(fiber)
}

func UseState[T any](ctx context.Context, initialValue T) (func() T, func(T)) {
	f := GetFiberFromContext(ctx)
	f.Mu.Lock()
	defer f.Mu.Unlock()

	if f.CallIndex >= len(f.StateOrder) {
		stateKey := len(f.StateOrder)
		f.StateOrder = append(f.StateOrder, stateKey)
		f.State[stateKey] = initialValue
	}

	stateKey := f.StateOrder[f.CallIndex]
	f.CallIndex++

	getState := func() T {
		f.Mu.Lock()
		defer f.Mu.Unlock()
		return f.State[stateKey].(T)
	}

	setState := func(newValue T) {
		f.Mu.Lock()
		f.State[stateKey] = newValue
		f.Mu.Unlock()
		root := f
		for root.Parent != nil {
			root = root.Parent
		}
		if root.Renderer != nil {
			root.Renderer.Schedule(root)
		}
	}

	return getState, setState
}

func UseEffect(ctx context.Context, setup func() func(), deps ...any) {
	f := GetFiberFromContext(ctx)
	f.Mu.Lock()
	defer f.Mu.Unlock()

	effectIndex := f.CallIndex
	f.CallIndex++

	if effectIndex < len(f.Effects) {
		// Subsequent render: check if dependencies changed
		oldDeps := f.Effects[effectIndex].Deps
		if !depsEqual(oldDeps, deps) {
			if f.Effects[effectIndex].Cleanup != nil {
				f.Effects[effectIndex].Cleanup()
			}
			cleanup := setup()
			f.Effects[effectIndex] = Effect{Setup: setup, Cleanup: cleanup, Deps: deps}
		}
	} else {
		// First render: run setup and store the effect
		cleanup := setup()
		f.Effects = append(f.Effects, Effect{Setup: setup, Cleanup: cleanup, Deps: deps})
	}
}

func depsEqual(oldDeps, newDeps []any) bool {
	if len(oldDeps) != len(newDeps) {
		return false
	}
	for i, old := range oldDeps {
		if old != newDeps[i] {
			return false
		}
	}
	return true
}

func runEffects(fiber *Fiber) {
	fiber.Mu.Lock()
	defer fiber.Mu.Unlock()
	for i, effect := range fiber.Effects {
		if effect.Cleanup != nil {
			effect.Cleanup()
		}
		cleanup := effect.Setup()
		fiber.Effects[i].Cleanup = cleanup
	}
}

func (r *Renderer) UnmountFiber(fiber *Fiber) {
	fiber.Mu.Lock()
	defer fiber.Mu.Unlock()
	for _, effect := range fiber.Effects {
		if effect.Cleanup != nil {
			effect.Cleanup()
		}
	}
	if fiber.DOM.Truthy() {
		fiber.Parent.DOM.Call("removeChild", fiber.DOM)
	}
}
