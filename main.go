package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strings"
)

const debug = true

//
// Compiler
//

type Compiler struct {
	directoryPath string
	filePaths     []string

	fileSet *token.FileSet
	files   []*ast.File
	types   *types.Info

	genTypeNames map[types.Type]string // Should we use `typeutil.Map` (equivalence-based keys)?
	genFuncDecls map[*ast.FuncDecl]string

	errors *strings.Builder
	output *strings.Builder
}

func (c *Compiler) errorf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(c.errors, "%s: ", c.fileSet.PositionFor(pos, true))
	fmt.Fprintf(c.errors, format, args...)
	fmt.Fprintln(c.errors)
}

func (c *Compiler) errored() bool {
	return c.errors.Len() != 0
}

func (c *Compiler) writef(format string, args ...interface{}) {
	fmt.Fprintf(c.output, format, args...)
}

func (c *Compiler) genTypeName(typ types.Type) string {
	if result, ok := c.genTypeNames[typ]; ok {
		return result
	} else {
		result = typ.String()
		c.genTypeNames[typ] = result
		return result
	}
}

func (c *Compiler) genFuncDecl(decl *ast.FuncDecl) string {
	if result, ok := c.genFuncDecls[decl]; ok {
		return result
	} else {
		signature := c.types.Defs[decl.Name].Type().(*types.Signature)
		results := signature.Results()
		if results.Len() > 1 {
			c.errorf(decl.Type.Results.Pos(), "multiple return values not supported")
		}

		builder := &strings.Builder{}

		// Return type
		if results.Len() == 0 {
			if decl.Name.String() == "main" && decl.Recv == nil {
				builder.WriteString("int ")
			} else {
				builder.WriteString("void ")
			}
		} else {
			builder.WriteString(c.genTypeName(results.At(0).Type()))
			builder.WriteByte(' ')
		}

		// Name
		builder.WriteString(decl.Name.String())

		// Parameters
		builder.WriteByte('(')
		builder.WriteByte(')')

		result = builder.String()
		c.genFuncDecls[decl] = result
		return result
	}
}

func (c *Compiler) writeFuncBody(decl *ast.FuncDecl) {
}

func (c *Compiler) compile() {
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
	c.genTypeNames = make(map[types.Type]string)
	c.genFuncDecls = make(map[*ast.FuncDecl]string)

	// Initialize builders
	c.errors = &strings.Builder{}
	c.output = &strings.Builder{}

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
	c.types = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	if _, err := (&types.Config{}).Check("", c.fileSet, c.files, c.types); err != nil {
		fmt.Println(err)
		return
	}

	// `#include`s
	c.writef("#include \"prelude.hh\"\n")

	// Function declarations
	c.writef("\n\n")
	for _, file := range c.files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				c.writef("%s;\n", c.genFuncDecl(decl))
			}
		}
	}

	// Function bodies
	c.writef("\n\n")
	for _, file := range c.files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				c.writef("%s {\n", c.genFuncDecl(decl))
				c.writeFuncBody(decl)
				c.writef("}\n")
			}
		}
	}

	// Debug
	if debug {
		fmt.Print("// types:")
		for typ := range c.genTypeNames {
			fmt.Printf(" %s", typ.String())
		}
		fmt.Print("\n\n")
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
		ioutil.WriteFile("output.cc", []byte(c.output.String()), fs.ModePerm)
	}
}
