package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
)

//
// Types
//

type Func struct {
	decl      *ast.FuncDecl
	signature *types.Signature

	outDecl string
}

type Compiler struct {
	directoryPath string
	filePaths     []string

	fileSet   *token.FileSet
	typesInfo *types.Info

	funcs []*Func

	outErrs *bytes.Buffer
	outCode *bytes.Buffer
}

//
// Errors
//

func (c *Compiler) eprintf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(c.outErrs, "%s: ", c.fileSet.PositionFor(pos, true))
	fmt.Fprintf(c.outErrs, format, args...)
	fmt.Fprintln(c.outErrs)
}

func (c *Compiler) errored() bool {
	return c.outErrs.Len() != 0
}

//
// Analysis
//

func (c *Compiler) analyzeFunc(decl *ast.FuncDecl) *Func {
	signature := c.typesInfo.Defs[decl.Name].Type().(*types.Signature)
	if signature.Results().Len() > 1 {
		c.eprintf(decl.Type.Results.Pos(), "multiple return values not supported")
	}

	//outputDeclBuf := &bytes.Buffer{}

	return &Func{
		decl:      decl,
		signature: signature,
		outDecl:   "void foo();",
	}
}

func (c *Compiler) analyze() {
	// Parse
	c.fileSet = token.NewFileSet()
	var files []*ast.File
	for _, filePath := range c.filePaths {
		file, err := parser.ParseFile(c.fileSet, filePath, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		files = append(files, file)
	}

	// Type-check
	config := types.Config{}
	c.typesInfo = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	if _, err := config.Check("", c.fileSet, files, c.typesInfo); err != nil {
		fmt.Println(err)
		return
	}

	// Functions
	for _, file := range files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				c.funcs = append(c.funcs, c.analyzeFunc(decl))
			}
		}
	}
}

//
// Writing
//

func (c *Compiler) writef(format string, args ...interface{}) {
	fmt.Fprintf(c.outCode, format, args...)
}

func (c *Compiler) write() {
	c.writef("#include \"../preamble.hh\"\n")

	c.writef("\n\n")
	for _, fun := range c.funcs {
		c.writef("%s\n", fun.outDecl)
	}
}

//
// Top-level
//

func (c *Compiler) init() {
	// Collect file paths from directory
	if len(c.filePaths) == 0 && c.directoryPath != "" {
		fileInfos, err := ioutil.ReadDir(c.directoryPath)
		if err != nil {
			fmt.Println(err)
			return
		}
		for _, fileInfo := range fileInfos {
			c.filePaths = append(c.filePaths, filepath.Join(c.directoryPath, fileInfo.Name()))
		}
	}

	// Initialize output buffers
	c.outErrs = &bytes.Buffer{}
	c.outCode = &bytes.Buffer{}
}

func (c *Compiler) compile() {
	c.init()
	c.analyze()
	if !c.errored() {
		c.write()
	}
}

//
// Main
//

func main() {
	// Compile
	c := Compiler{directoryPath: "example"}
	c.compile()

	// Print output
	if c.errored() {
		fmt.Println(c.outErrs)
	} else {
		fmt.Println(c.outCode)
	}
}
