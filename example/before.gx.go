// This file appears before 'main.gx.go' in AST order. Used to test that
// references to symbols from later files leads to proper hoisting in output.

package main

type Before struct {
	p Point // Reference to type spec from main
}

var globalW = globalX + 42 // Reference to value spec from main
