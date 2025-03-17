package goat

import "syscall/js"

type VNode struct {
	Tag       string
	Attrs     map[string]string
	Events    map[string]js.Func
	Children  []VNode
	Text      string
	Key       string
	Component Component
	Props     any
}
