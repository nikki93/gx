package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
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

	typeCppNames map[types.Type]string // Should we use `typeutil.Map`?
	funcCppDecls map[*ast.FuncDecl]string

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

func (c *Compiler) typeCppName(typ types.Type) string {
	if result, ok := c.typeCppNames[typ]; ok {
		return result
	} else {
		result = typ.String()
		c.typeCppNames[typ] = result
		return result
	}
}

func (c *Compiler) funcCppDecl(decl *ast.FuncDecl) string {
	if result, ok := c.funcCppDecls[decl]; ok {
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
			builder.WriteString("void ")
		} else {
			builder.WriteString(c.typeCppName(results.At(0).Type()))
			builder.WriteByte(' ')
		}

		// Name
		builder.WriteString(decl.Name.String())

		// Parameters

		result = builder.String()
		c.funcCppDecls[decl] = result
		return result
	}
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
	c.typeCppNames = make(map[types.Type]string)
	c.funcCppDecls = make(map[*ast.FuncDecl]string)

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

	// Write preamble
	c.writef("#include \"../preamble.hh\"\n")

	// Write function declarations
	c.writef("\n\n")
	for _, file := range c.files {
		for _, decl := range file.Decls {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				c.writef("%s;\n", c.funcCppDecl(decl))
			}
		}
	}

	// Debug
	if debug {
		fmt.Println("types: ")
		for typ := range c.typeCppNames {
			fmt.Printf("  %s\n", typ.String())
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
	}
}
