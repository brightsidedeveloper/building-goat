package goat

import (
	"context"
	"fmt"
	"sync"
	"syscall/js"
	"time"
)

type ComponentInstance struct {
	states     map[int]any
	callbacks  map[int]string
	jsFuncs    map[int]js.Func
	stateOrder []int
	callIndex  int
	mu         sync.Mutex
}

type componentInstanceKeyType struct{}

var componentInstanceKey = componentInstanceKeyType{}

func getInstanceFromContext(ctx context.Context) *ComponentInstance {
	if ci, ok := ctx.Value(componentInstanceKey).(*ComponentInstance); ok {
		return ci
	}
	panic("No component instance found in context")
}

func UseState[T any](ctx context.Context, initialValue T) (func() T, func(T)) {
	ci := getInstanceFromContext(ctx)
	ci.mu.Lock()
	defer ci.mu.Unlock()

	if ci.callIndex >= len(ci.stateOrder) {
		stateKey := len(ci.stateOrder)
		ci.stateOrder = append(ci.stateOrder, stateKey)
		ci.states[stateKey] = initialValue
	}

	stateKey := ci.stateOrder[ci.callIndex]
	ci.callIndex++

	getState := func() T {
		ci.mu.Lock()
		defer ci.mu.Unlock()
		return ci.states[stateKey].(T)
	}

	setState := func(newValue T) {
		ci.mu.Lock()
		ci.states[stateKey] = newValue
		ci.mu.Unlock()
		if renderer := getRendererForInstance(ci); renderer != nil {
			go renderer.Render()
		}
	}

	return getState, setState
}

func UseCallback(ctx context.Context, f func(this js.Value, args []js.Value) any) js.Func {
	ci := getInstanceFromContext(ctx)
	ci.mu.Lock()
	defer ci.mu.Unlock()

	callbackIndex := ci.callIndex
	ci.callIndex++

	if oldName, exists := ci.callbacks[callbackIndex]; exists {
		js.Global().Delete(oldName)
		if oldFunc, ok := ci.jsFuncs[callbackIndex]; ok {
			oldFunc.Release() // Clean up old js.Func
		}
	}

	name := fmt.Sprintf("fn%d_%d", time.Now().UnixNano(), callbackIndex)
	jsFunc := js.FuncOf(f)
	js.Global().Set(name, jsFunc)
	ci.callbacks[callbackIndex] = name

	if ci.jsFuncs == nil {
		ci.jsFuncs = make(map[int]js.Func)
	}
	ci.jsFuncs[callbackIndex] = jsFunc

	Log("UseCallback created:", name) // Debug
	return jsFunc
}

func UseEffect(ctx context.Context, effect func()) {
	ci := getInstanceFromContext(ctx)
	ci.mu.Lock()
	defer ci.mu.Unlock()
	effect()
}

func (ci *ComponentInstance) Reset() {
	ci.mu.Lock()
	defer ci.mu.Unlock()
	ci.callIndex = 0
}
