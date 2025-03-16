package goat

import (
	"sync"
)

var (
	renderersMu sync.Mutex
	renderers   = make(map[string]*Renderer)
)

func RenderRoot(id string, comp Component, props any) {
	renderersMu.Lock()
	defer renderersMu.Unlock()
	r := NewRenderer(id, comp, props)
	renderers[id] = r
	r.Render()
}
