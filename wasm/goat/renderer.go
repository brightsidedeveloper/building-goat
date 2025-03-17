package goat

import (
	"reflect"
	"sync"
	"syscall/js"
)

type Renderer struct {
	RootFiber *Fiber
	DOMID     string
	Mu        sync.Mutex
	WorkQueue chan *Fiber
}

func NewRenderer(id string, comp Component, props any) *Renderer {
	r := &Renderer{
		RootFiber: NewFiber(comp, props, nil, ""),
		DOMID:     id,
		WorkQueue: make(chan *Fiber, 100),
	}
	r.RootFiber.Renderer = r
	go r.WorkLoop()
	return r
}

func RenderRoot(id string, comp Component, props any) {
	r := NewRenderer(id, comp, props)
	r.Schedule(r.RootFiber)
	select {}
}

func (r *Renderer) Schedule(fiber *Fiber) {
	fiber.Dirty = true
	r.WorkQueue <- fiber
}

func (r *Renderer) WorkLoop() {
	for fiber := range r.WorkQueue {
		r.RenderFiber(fiber)
		r.CommitFiber(fiber)
	}
}

func (r *Renderer) ReconcileChildren(fiber *Fiber) {
	oldChild := fiber.Alternate != nil && fiber.Alternate.Child != nil
	if fiber.Component != nil && !reflect.ValueOf(fiber.Node).IsZero() {
		var childFiber *Fiber
		if oldChild {
			childFiber = fiber.Alternate.Child
			childFiber.Node = fiber.Node
			childFiber.Dirty = true
		} else {
			childFiber = NewFiber(nil, nil, fiber, fiber.Node.Key)
			childFiber.Node = fiber.Node
		}
		fiber.Child = childFiber
		r.reconcileHostChildren(childFiber)
	} else {
		r.reconcileHostChildren(fiber)
	}
}

func (r *Renderer) reconcileHostChildren(fiber *Fiber) {
	oldChild := fiber.Alternate != nil && fiber.Alternate.Child != nil
	var prevFiber *Fiber
	oldByKey := make(map[string]*Fiber)
	if oldChild {
		for child := fiber.Alternate.Child; child != nil; child = child.Sibling {
			if child.Key != "" {
				oldByKey[child.Key] = child
			}
		}
	}
	newChildren := fiber.Node.Children
	for i, childNode := range newChildren {
		var childFiber *Fiber
		oldFiber := oldByKey[childNode.Key]
		if childNode.Component != nil {
			if oldFiber != nil && oldFiber.Component != nil && oldFiber.Key == childNode.Key {
				childFiber = oldFiber
				childFiber.Props = childNode.Props
				childFiber.Dirty = true
			} else {
				childFiber = NewFiber(childNode.Component, childNode.Props, fiber, childNode.Key)
			}
		} else {
			if oldFiber != nil && oldFiber.Component == nil && oldFiber.Key == childNode.Key {
				childFiber = oldFiber
				childFiber.Node = childNode
				childFiber.Dirty = true
			} else {
				childFiber = NewFiber(nil, nil, fiber, childNode.Key)
				childFiber.Node = childNode
			}
		}
		if i == 0 {
			fiber.Child = childFiber
		} else if prevFiber != nil {
			prevFiber.Sibling = childFiber
		}
		prevFiber = childFiber
		if childNode.Key != "" {
			delete(oldByKey, childNode.Key)
		}
	}
	for _, old := range oldByKey {
		r.UnmountFiber(old)
	}
}

func (r *Renderer) CommitFiber(fiber *Fiber) {
	doc := js.Global().Get("document")
	container := doc.Call("getElementById", r.DOMID)
	if !container.Truthy() {
		return
	}
	if fiber.Parent == nil {
		if fiber.Child != nil {
			diffAndPatch(fiber.Child, container, doc)
		}
	} else {
		diffAndPatch(fiber, fiber.Parent.DOM, doc)
	}
	runEffects(fiber)
}

func diffAndPatch(fiber *Fiber, parentDOM js.Value, doc js.Value) {
	if fiber.Component != nil {
		fiber.Renderer.RenderFiber(fiber)
		if fiber.Child != nil {
			diffAndPatch(fiber.Child, parentDOM, doc)
		}
	} else {
		if fiber.DOM.IsUndefined() {
			fiber.DOM = createDOM(fiber.Node, doc)
			parentDOM.Call("appendChild", fiber.DOM)
		} else if fiber.Alternate != nil {
			updateAttributesAndEvents(fiber.DOM, fiber.Alternate.Node, fiber.Node)
		}
		child := fiber.Child
		for child != nil {
			diffAndPatch(child, fiber.DOM, doc)
			child = child.Sibling
		}
	}
}

func updateAttributesAndEvents(dom js.Value, oldNode, newNode VNode) {
	for k, v := range newNode.Attrs {
		if oldVal, ok := oldNode.Attrs[k]; !ok || oldVal != v {
			dom.Call("setAttribute", k, v)
		}
	}
	for k := range oldNode.Attrs {
		if _, ok := newNode.Attrs[k]; !ok {
			dom.Call("removeAttribute", k)
		}
	}
	for e, h := range newNode.Events {
		if oldHandler, ok := oldNode.Events[e]; ok {
			if !oldHandler.Value.Equal(h.Value) {
				dom.Call("removeEventListener", e, oldHandler)
				dom.Call("addEventListener", e, h)
			}
		} else {
			dom.Call("addEventListener", e, h)
		}
	}
	for e, oldHandler := range oldNode.Events {
		if _, ok := newNode.Events[e]; !ok {
			dom.Call("removeEventListener", e, oldHandler)
		}
	}
}

func createDOM(node VNode, doc js.Value) js.Value {
	if node.Tag == "" && node.Text != "" {
		return doc.Call("createTextNode", node.Text)
	}
	if node.Tag != "" && node.Text != "" {
		js.Global().Get("console").Call("log", "VNode with Tag should not have Text; use Children instead: "+node.Tag)
	}
	elem := doc.Call("createElement", node.Tag)
	for k, v := range node.Attrs {
		elem.Call("setAttribute", k, v)
	}
	for e, h := range node.Events {
		elem.Call("addEventListener", e, h)
	}
	for _, child := range node.Children {
		elem.Call("appendChild", createDOM(child, doc))
	}
	return elem
}
