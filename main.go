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
// Types
//

type Func struct {
	decl      *ast.FuncDecl
	signature *types.Signature

	prototype string
}

type Compiler struct {
	inputFilename string
	fileSet       *token.FileSet

	typesInfo *types.Info

	funcs []*Func

	errors *bytes.Buffer
	output *bytes.Buffer
}

//
// Errors
//

func (comp *Compiler) eprintf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(comp.errors, "%s: ", comp.fileSet.PositionFor(pos, true))
	fmt.Fprintf(comp.errors, format, args...)
	fmt.Fprintln(comp.errors)
}

func (comp *Compiler) errored() bool {
	return comp.errors.Len() != 0
}

//
// Functions
//

func (comp *Compiler) makeFunc(decl *ast.FuncDecl) *Func {
	signature := comp.typesInfo.Defs[decl.Name].Type().(*types.Signature)
	if signature.Results().Len() > 1 {
		comp.eprintf(decl.Type.Results.Pos(), "multiple return values not supported")
	}

	//prototypeBuf := &bytes.Buffer{}

	return &Func{
		decl:      decl,
		signature: signature,
		prototype: "void foo();",
	}
}

//
// Top-level
//

func (comp *Compiler) analyze() {
	// Initialize
	comp.errors = &bytes.Buffer{}
	comp.output = &bytes.Buffer{}

	// Parse
	comp.fileSet = token.NewFileSet()
	astFile, err := parser.ParseFile(comp.fileSet, comp.inputFilename, nil, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Type-check
	config := types.Config{}
	comp.typesInfo = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	_, err = config.Check("", comp.fileSet, []*ast.File{astFile}, comp.typesInfo)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Functions
	for _, decl := range astFile.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			comp.funcs = append(comp.funcs, comp.makeFunc(decl))
		}
	}
}

func (comp *Compiler) writeSectionComment(sectionName string) {
	fmt.Fprintf(comp.output, "\n\n//\n// %s\n//\n\n", sectionName)
}

func (comp *Compiler) write() {
	// Preamble
	fmt.Fprintf(comp.output, "#include \"../preamble.hh\"\n")

	// Function prototypes
	comp.writeSectionComment("Function prototypes")
	for _, fun := range comp.funcs {
		fmt.Fprintln(comp.output, fun.prototype)
	}
}

func (comp *Compiler) compile() {
	comp.analyze()
	if !comp.errored() {
		comp.write()
	}
}

//
// Main
//

func main() {
	// Compile
	comp := Compiler{inputFilename: "examples/basic_1.go"}
	comp.compile()

	// Print output
	if comp.errored() {
		fmt.Println(comp.errors)
	} else {
		fmt.Println(comp.output)
	}
}
