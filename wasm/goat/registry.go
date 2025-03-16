package goat

import "sync"

var rendererRegistry = struct {
	sync.Mutex
	m map[*ComponentInstance]*Renderer
}{m: make(map[*ComponentInstance]*Renderer)}

func registerRenderer(ci *ComponentInstance, r *Renderer) {
	rendererRegistry.Lock()
	defer rendererRegistry.Unlock()
	rendererRegistry.m[ci] = r
}

func getRendererForInstance(ci *ComponentInstance) *Renderer {
	rendererRegistry.Lock()
	defer rendererRegistry.Unlock()
	return rendererRegistry.m[ci]
}
