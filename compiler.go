package main

import (
	"log"
	"github.com/mini-compiler/checkers"
	"strings"
	"fmt"
)

// TOKENIZER/LEXER

type token struct {
	kind  string
	value string
}

func tokenizer(input string) []token {
	input += "\n"
	current := 0
	var tokens []token

	for current < len(input) {
		char := string(input[current])

		switch char {
		case "(":
			tokens = append(tokens, token{
				kind:  "paren",
				value: "(",
			})
			current++
			continue
		case ")":
			tokens = append(tokens, token{
				kind:  "paren",
				value: ")",
			})
			current++
			continue
		case " ":
			current++
			continue
		default:
		}

		if checkers.IsNumber(char) {
			value := ""

			for checkers.IsNumber(char) {
				value += char
				current++
				char = string(input[current])
			}

			tokens = append(tokens, token{
				kind:  "number",
				value: value,
			})
			continue
		}

		if checkers.IsLetter(char) {
			value := ""

			for checkers.IsLetter(char) {
				value += char
				current++
				char = string(input[current])
			}

			tokens = append(tokens, token{
				kind:  "name",
				value: value,
			})
			continue
		}
		break
	}
	return tokens
}

// PARSER

// AST - Abstract syntax tree
// [{ type: 'paren', value: '(' }, ...]   =>   { type: 'Program', body: [...] }

type node struct {
	kind       string
	value      string
	name       string
	callee     *node
	expression *node
	body       []node
	params     []node
	arguments  *[]node
	context    *[]node
}

type ast node

var parsingCounter int
var pTokens []token

func parser(tokens []token) ast {
	pTokens = tokens

	ast := ast{
		kind: "Program",
		body: []node{},
	}

	parsingCounter = 0
	for parsingCounter < len(pTokens) {
		ast.body = append(ast.body, walk())
	}

	return ast
}

func walk() node {
	token := pTokens[parsingCounter]
	if token.kind == "number" {
		parsingCounter++
		return node{
			kind:  "NumberLiteral",
			value: token.value,
		}
	}

	if token.kind == "paren" && token.value == "(" {
		parsingCounter++
		token = pTokens[parsingCounter]
		n := node{
			kind:   "CallExpression",
			name:   token.value,
			params: []node{},
		}
		parsingCounter++
		token = pTokens[parsingCounter]

		for token.kind != "paren" || (token.kind == "paren" && token.value != ")") {
			n.params = append(n.params, walk())
			token = pTokens[parsingCounter]
		}
		parsingCounter++

		return n
	}

	log.Fatal(token.kind)
	return node{}
}

type visitor map[string]func(n *node, p node)

func traverser(a ast, v visitor) {
	traverseNode(node(a), node{}, v)
}

func traverseArray(a []node, p node, v visitor) {
	for _, child := range a {
		traverseNode(child, p, v)
	}
}

func traverseNode(n, p node, v visitor) {
	for k, va := range v {
		if k == n.kind {
			va(&n, p)
		}
	}
	switch n.kind {

	case "Program":
		traverseArray(n.body, n, v)
		break

	case "CallExpression":
		traverseArray(n.params, n, v)
		break

	case "NumberLiteral":
		break

	default:
		log.Fatal(n.kind)
	}
}


func transformer(a ast) ast {
	newAst := ast{
		kind: "Program",
		body: []node{},
	}
	a.context = &newAst.body

	traverser(a, map[string]func(n *node, p node){
		"NumberLiteral": func(n *node, p node) {
			*p.context = append(*p.context, node{
				kind:  "NumberLiteral",
				value: n.value,
			})
		},

		"CallExpression": func(n *node, p node) {

			e := node{
				kind: "CallExpression",
				callee: &node{
					kind: "Identifier",
					name: n.name,
				},
				arguments: new([]node),
			}
			n.context = e.arguments


			if p.kind != "CallExpression" {
				es := node{
					kind:       "ExpressionStatement",
					expression: &e,
				}
				*p.context = append(*p.context, es)
			} else {
				*p.context = append(*p.context, e)
			}

		},
	})
	return newAst
}

//CODE GENERATOR

func codeGenerator(n node) string {
	switch n.kind {

	case "Program":
		var r []string
		for _, no := range n.body {
			r = append(r, codeGenerator(no))
		}
		return strings.Join(r, "\n")

	case "ExpressionStatement":
		return codeGenerator(*n.expression) + ";"

	case "CallExpression":
		var ra []string
		c := codeGenerator(*n.callee)

		for _, no := range *n.arguments {
			ra = append(ra, codeGenerator(no))
		}

		r := strings.Join(ra, ", ")
		return c + "(" + r + ")"

	case "Identifier":
		return n.name

	case "NumberLiteral":
		return n.value

	default:
		log.Fatal("err")
		return ""
	}
}

func compiler(input string) string {
	tokens := tokenizer(input)
	ast := parser(tokens)
	nast := transformer(ast)
	out := codeGenerator(node(nast))
	return out
}

func main() {
	program := ""
	program += "(add 10 5)"
	program += "(add 10 (asd 10 6))"
	program += "(add 10 (qwerty (asd 10 4) 6))"
	program += "(add 10 (subtract (quote 5 4) 6))"
	program += "(add 10 (subtract (quote (saver 3 4) 4) 6))"
	out := compiler(program)
	fmt.Println(out)
}
