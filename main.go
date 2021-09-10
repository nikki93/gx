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

type Compiler struct {
	directoryPath string
	filePaths     []string

	fileSet *token.FileSet
	files   []*ast.File
	types   *types.Info

	funcCppDecls map[*ast.FuncDecl]string

	errors *bytes.Buffer
	output *bytes.Buffer
}

//
// Errors
//

func (c *Compiler) eprintf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(c.errors, "%s: ", c.fileSet.PositionFor(pos, true))
	fmt.Fprintf(c.errors, format, args...)
	fmt.Fprintln(c.errors)
}

func (c *Compiler) errored() bool {
	return c.errors.Len() != 0
}

//
// Analysis
//

func (c *Compiler) funcCppDecl(decl *ast.FuncDecl) string {
	if result, ok := c.funcCppDecls[decl]; ok {
		return result
	} else {
		//signature := c.types.Defs[funcDecl.Name].Type().(*types.Signature)
		result = "void foo();"
		c.funcCppDecls[decl] = result
		return result
	}
}

func (c *Compiler) analyze() {
	// Parse
	c.fileSet = token.NewFileSet()
	for _, filePath := range c.filePaths {
		file, err := parser.ParseFile(c.fileSet, filePath, nil, 0)
		if err != nil {
			fmt.Println(err)
			return
		}
		c.files = append(c.files, file)
	}

	// Type-check
	config := types.Config{}
	c.types = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	if _, err := config.Check("", c.fileSet, c.files, c.types); err != nil {
		fmt.Println(err)
		return
	}
}

//
// Writing
//

func (c *Compiler) writef(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
}

func (c *Compiler) write() {
	// Preamble
	c.writef("#include \"../preamble.hh\"\n")

	// Function declarations
	c.writef("\n\n")
	for _, file := range c.files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				c.writef("%s\n", c.funcCppDecl(decl))
			}
		}
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

	// Initialize maps
	c.funcCppDecls = make(map[*ast.FuncDecl]string)

	// Initialize buffers
	c.errors = &bytes.Buffer{}
	c.output = &bytes.Buffer{}
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
		fmt.Println(c.errors)
	} else {
		fmt.Println(c.output)
	}
}
