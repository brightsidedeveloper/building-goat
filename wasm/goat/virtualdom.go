package goat

import "syscall/js"

type GoatNode struct {
	Tag      string
	Attrs    map[string]string
	Events   map[string]js.Func
	Children []GoatNode
	Text     string
	DOMNode  js.Value
}

func NewGoatNode(tag string, attrs map[string]string, events map[string]js.Func, children []GoatNode, text string) GoatNode {
	return GoatNode{
		Tag:      tag,
		Attrs:    attrs,
		Events:   events,
		Children: children,
		Text:     text,
	}
}
