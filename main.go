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

func unimpl() {
	panic("unimplemented")
}

//
// Compiler
//

type Compiler struct {
	mainFilename string

	info *types.Info
	pkg  *types.Package

	topTypes []*ast.TypeSpec
	topFuncs []*ast.FuncDecl

	result bytes.Buffer
}

// Top-level

func (comp *Compiler) compile() {
	// Parse
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, comp.mainFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
	}

	// Type-check
	config := &types.Config{}
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
	}

	// Collect top-level decls
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			switch decl.Tok {
			case token.CONST:
			case token.VAR:
				for _, spec := range decl.Specs {
					for _, name := range spec.(*ast.ValueSpec).Names {
						fmt.Printf("%s: global variable not allowed: %s\n", fileSet.PositionFor(spec.Pos(), true), name)
					}
				}
			case token.TYPE:
				for _, spec := range decl.Specs {
					comp.topTypes = append(comp.topTypes, spec.(*ast.TypeSpec))
				}
			}
		case *ast.FuncDecl:
			comp.topFuncs = append(comp.topFuncs, decl)
		}
	}

	// Debug
	for _, spec := range comp.topTypes {
		fmt.Println("type: ", spec.Name)
	}
	for _, decl := range comp.topFuncs {
		fmt.Println("func: ", decl.Name)
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
