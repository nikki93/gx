package main

import (
	"fmt"
	"go/parser"
	"go/token"
)

func main() {
	fmt.Println("welcome to the gx compiler ;D")
	fset := token.NewFileSet()
	root, err := parser.ParseFile(fset, "examples/basic_1.go", nil, 0)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(root)
}
