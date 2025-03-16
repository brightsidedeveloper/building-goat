package goat

import (
	"context"
	"sync"
	"syscall/js"
)

type Renderer struct {
	instance *ComponentInstance
	comp     Component
	props    any
	id       string
	vdom     GoatNode
	mu       sync.Mutex
}

func NewRenderer(id string, comp Component, props any) *Renderer {
	instance := &ComponentInstance{
		states:     make(map[int]any),
		stateOrder: []int{},
		callbacks:  make(map[int]string),
		callIndex:  0,
	}
	r := &Renderer{
		instance: instance,
		comp:     comp,
		props:    props,
		id:       id,
	}
	registerRenderer(instance, r)
	return r
}

func (r *Renderer) Render() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instance.Reset()
	ctx := context.WithValue(context.Background(), componentInstanceKey, r.instance)
	ctx = context.WithValue(ctx, propsKey, r.props)
	newVdom := r.comp(ctx, r.props)
	r.updateDOM(newVdom)
}

func (r *Renderer) updateDOM(newVdom GoatNode) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", r.id)
	if !container.Truthy() {
		return
	}

	if !r.vdom.DOMNode.Truthy() {
		r.vdom = createDOM(newVdom, doc)
		container.Call("appendChild", r.vdom.DOMNode)
	} else {
		diffAndPatch(&r.vdom, &newVdom, container, doc)
	}
}

func createDOM(v GoatNode, doc js.Value) GoatNode {
	var node js.Value
	if v.Tag == "" && v.Text != "" && len(v.Children) == 0 {
		node = doc.Call("createTextNode", v.Text)
	} else if v.Tag != "" {
		node = doc.Call("createElement", v.Tag)
		for key, value := range v.Attrs {
			node.Call("setAttribute", key, value)
		}
		for event, handler := range v.Events {
			node.Call("addEventListener", event, handler)
		}
		children := make([]GoatNode, len(v.Children))
		for i, child := range v.Children {
			childNode := createDOM(child, doc)
			node.Call("appendChild", childNode.DOMNode)
			children[i] = childNode
		}
		v.Children = children
		if len(v.Children) == 0 && v.Text != "" {
			node.Set("textContent", v.Text)
		}
	}
	v.DOMNode = node
	return v
}

func diffAndPatch(oldVdom, newVdom *GoatNode, parent js.Value, doc js.Value) {
	if !parent.Truthy() {
		return
	}

	if oldVdom.Tag != newVdom.Tag || (oldVdom.Tag == "" && oldVdom.Text != newVdom.Text) {
		newNode := createDOM(*newVdom, doc)
		if !oldVdom.DOMNode.Truthy() {
			parent.Call("appendChild", newNode.DOMNode)
		} else {
			parent.Call("replaceChild", newNode.DOMNode, oldVdom.DOMNode)
		}
		*oldVdom = newNode // Replace entire node, including Text
		return
	}

	if oldVdom.Tag != "" {
		if len(oldVdom.Children) == 0 && oldVdom.Text != newVdom.Text {
			oldVdom.DOMNode.Set("textContent", newVdom.Text)
			oldVdom.Text = newVdom.Text // Sync Text
		}

		for key, value := range newVdom.Attrs {
			if oldVdom.Attrs[key] != value {
				oldVdom.DOMNode.Call("setAttribute", key, value)
			}
		}
		for key := range oldVdom.Attrs {
			if _, exists := newVdom.Attrs[key]; !exists {
				oldVdom.DOMNode.Call("removeAttribute", key)
			}
		}
		oldVdom.Attrs = newVdom.Attrs // Sync Attrs

		for event, oldHandler := range oldVdom.Events {
			oldVdom.DOMNode.Call("removeEventListener", event, oldHandler)
		}
		for event, handler := range newVdom.Events {
			oldVdom.DOMNode.Call("addEventListener", event, handler)
			oldVdom.Events[event] = handler
		}
		oldVdom.Events = newVdom.Events // Sync Events

		oldLen := len(oldVdom.Children)
		newLen := len(newVdom.Children)
		for i := 0; i < oldLen || i < newLen; i++ {
			if i >= oldLen && i < newLen {
				newChild := createDOM(newVdom.Children[i], doc)
				oldVdom.DOMNode.Call("appendChild", newChild.DOMNode)
				oldVdom.Children = append(oldVdom.Children, newChild)
			} else if i < oldLen && i >= newLen {
				oldVdom.DOMNode.Call("removeChild", oldVdom.Children[i].DOMNode)
				oldVdom.Children = oldVdom.Children[:i]
				break
			} else {
				diffAndPatch(&oldVdom.Children[i], &newVdom.Children[i], oldVdom.DOMNode, doc)
			}
		}
	}
}

type propsKeyType struct{}

var propsKey = propsKeyType{}

func GetProps(ctx context.Context) any {
	if props, ok := ctx.Value(propsKey).(any); ok {
		return props
	}
	panic("No props found in context")
}
