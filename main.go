package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
)

//
// Errors
//

func unimplemented() {
	panic("unimplemented")
}

//
// Compiler
//

type Compiler struct {
	mainFilename string

	info *types.Info
	pkg  *types.Package

	result bytes.Buffer
}

// Top-level

func (comp *Compiler) compile() {
	// Parse
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, comp.mainFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Type-check
	config := types.Config{}
	comp.info = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	comp.pkg, err = config.Check("", fileSet, []*ast.File{file}, comp.info)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Compile functions
	var funcDecls []*ast.FuncDecl
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			funcDecls = append(funcDecls, decl)
		}
	}
	for _, decl := range funcDecls {
		fmt.Printf("func: %s\n", decl.Name)
	}
}

//
// Main
//

func main() {
	// Compile
	comp := Compiler{mainFilename: "examples/basic_1.go"}
	comp.compile()

	// Print
	fmt.Println(comp.result.String())
}
