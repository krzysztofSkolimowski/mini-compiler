package main

type Node struct {
	kind       string
	value      string
	name       string
	callee     *Node
	expression *Node
	body       []Node
	params     []Node
	arguments  *[]Node
	context    *[]Node
}

type Ast Node
