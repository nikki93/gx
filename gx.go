package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/packages"
)

type Compiler struct {
	mainPkgPath string

	fileSet *token.FileSet
	types   *types.Info

	externs      map[types.Object]string
	fieldIndices map[*types.Var]int
	genTypeExprs map[types.Type]string
	genTypeDecls map[*ast.TypeSpec]string
	genTypeDefns map[*ast.TypeSpec]string
	genTypeMetas map[*ast.TypeSpec]string
	genFuncDecls map[*ast.FuncDecl]string

	indent     int
	errors     *strings.Builder
	outputCC   *strings.Builder
	outputHH   *strings.Builder
	atBlockEnd bool
}

//
// Error and writing utilities
//

func (c *Compiler) errorf(pos token.Pos, format string, args ...interface{}) {
	fmt.Fprintf(c.errors, "%s: ", c.fileSet.PositionFor(pos, true))
	fmt.Fprintf(c.errors, format, args...)
	fmt.Fprintln(c.errors)
}

func (c *Compiler) errored() bool {
	return c.errors.Len() != 0
}

func (c *Compiler) write(s string) {
	c.atBlockEnd = false
	if peek := c.outputCC.String(); len(peek) > 0 && peek[len(peek)-1] == '\n' {
		for i := 0; i < 2*c.indent; i++ {
			c.outputCC.WriteByte(' ')
		}
	}
	c.outputCC.WriteString(s)
}

func trimFinalSpace(s string) string {
	if l := len(s); l > 0 && s[l-1] == ' ' {
		return s[0 : l-1]
	} else {
		return s
	}
}

func lowerFirst(s string) string {
	result := []rune(s)
	result[0] = unicode.ToLower(result[0])
	return string(result)
}

//
// Types
//

func (c *Compiler) computeFieldIndices(typ *types.Struct) {
	nFields := typ.NumFields()
	if nFields == 0 {
		return
	}
	if _, ok := c.fieldIndices[typ.Field(0)]; ok {
		return
	}
	for i := 0; i < nFields; i++ {
		c.fieldIndices[typ.Field(i)] = i
	}
}

func (c *Compiler) genTypeExpr(typ types.Type, pos token.Pos) string {
	if result, ok := c.genTypeExprs[typ]; ok {
		return result
	} else {
		builder := &strings.Builder{}
		switch typ := typ.(type) {
		case *types.Basic:
			switch typ.Kind() {
			case types.Bool:
				builder.WriteString("bool")
			case types.Int:
				builder.WriteString("int")
			case types.Float32:
				builder.WriteString("float")
			case types.Float64:
				builder.WriteString("double")
			case types.Byte:
				builder.WriteString("unsigned char")
			case types.String:
				builder.WriteString("gx::String")
			default:
				c.errorf(pos, "%s not supported", typ.String())
			}
			builder.WriteByte(' ')
		case *types.Pointer:
			builder.WriteString(c.genTypeExpr(typ.Elem(), pos))
			builder.WriteByte('*')
		case *types.Named:
			name := typ.Obj()
			if ext, ok := c.externs[name]; ok {
				builder.WriteString(ext)
			} else {
				builder.WriteString(name.Name())
			}
			if typeArgs := typ.TypeArgs(); typeArgs != nil {
				builder.WriteString("<")
				for i, nTypeArgs := 0, typeArgs.Len(); i < nTypeArgs; i++ {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(trimFinalSpace(c.genTypeExpr(typeArgs.At(i), pos)))
				}
				builder.WriteString(">")
			}
			builder.WriteByte(' ')
		case *types.TypeParam:
			builder.WriteString(typ.Obj().Name())
			builder.WriteByte(' ')
		case *types.Array:
			builder.WriteString("gx::Array<")
			builder.WriteString(trimFinalSpace(c.genTypeExpr(typ.Elem(), pos)))
			builder.WriteString(", ")
			builder.WriteString(strconv.FormatInt(typ.Len(), 10))
			builder.WriteString(">")
			builder.WriteByte(' ')
		case *types.Slice:
			builder.WriteString("gx::Slice<")
			builder.WriteString(trimFinalSpace(c.genTypeExpr(typ.Elem(), pos)))
			builder.WriteString(">")
			builder.WriteByte(' ')
		default:
			c.errorf(pos, "%s not supported", typ.String())
		}
		result = builder.String()
		c.genTypeExprs[typ] = result
		return result
	}
}

func (c *Compiler) genTypeDecl(typeSpec *ast.TypeSpec) string {
	if result, ok := c.genTypeDecls[typeSpec]; ok {
		return result
	} else {
		builder := &strings.Builder{}
		if typeSpec.TypeParams != nil {
			builder.WriteString("template<")
			for i, typeParam := range typeSpec.TypeParams.List {
				for j, name := range typeParam.Names {
					if i > 0 || j > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString("typename ")
					builder.WriteString(name.String())
				}
			}
			builder.WriteString(">\n")
		}
		switch typeSpec.Type.(type) {
		case *ast.StructType:
			builder.WriteString("struct ")
			builder.WriteString(typeSpec.Name.String())
		case *ast.InterfaceType:
			// Empty -- only used as generic constraint during typecheck
			builder = &strings.Builder{}
		default:
			builder.WriteString("using ")
			builder.WriteString(typeSpec.Name.String())
			builder.WriteString(" = ")
			typ := c.types.TypeOf(typeSpec.Type)
			builder.WriteString(trimFinalSpace(c.genTypeExpr(typ, typeSpec.Type.Pos())))
		}
		result = builder.String()
		c.genTypeDecls[typeSpec] = result
		return result
	}
}

func (c *Compiler) genTypeDefn(typeSpec *ast.TypeSpec) string {
	if result, ok := c.genTypeDefns[typeSpec]; ok {
		return result
	} else {
		builder := &strings.Builder{}
		switch typ := typeSpec.Type.(type) {
		case *ast.StructType:
			builder.WriteString(c.genTypeDecl(typeSpec))
			builder.WriteString(" {\n")
			for _, field := range typ.Fields.List {
				if fieldType := c.types.TypeOf(field.Type); fieldType != nil {
					var defaultVal string
					if tag := field.Tag; tag != nil && tag.Kind == token.STRING {
						unquoted, _ := strconv.Unquote(tag.Value)
						defaultVal = reflect.StructTag(unquoted).Get("default")
					}
					typeExpr := c.genTypeExpr(fieldType, field.Type.Pos())
					for _, fieldName := range field.Names {
						builder.WriteString("  ")
						builder.WriteString(typeExpr)
						builder.WriteString(fieldName.String())
						if defaultVal != "" {
							builder.WriteString(" = ")
							builder.WriteString(defaultVal)
						}
						builder.WriteString(";\n")
					}
				}
			}
			builder.WriteByte('}')
		case *ast.InterfaceType:
			// Empty -- only used as generic constraint during typecheck
		default:
			// Empty -- alias declaration is definition
		}
		result = builder.String()
		c.genTypeDefns[typeSpec] = result
		return result
	}
}

func (c *Compiler) genTypeMeta(typeSpec *ast.TypeSpec) string {
	if result, ok := c.genTypeMetas[typeSpec]; ok {
		return result
	} else {
		builder := &strings.Builder{}
		switch typ := typeSpec.Type.(type) {
		case *ast.StructType:
			typeParamsBuilder := &strings.Builder{}
			if typeSpec.TypeParams != nil {
				for i, typeParam := range typeSpec.TypeParams.List {
					for j, name := range typeParam.Names {
						if i > 0 || j > 0 {
							typeParamsBuilder.WriteString(", ")
						}
						typeParamsBuilder.WriteString("typename ")
						typeParamsBuilder.WriteString(name.String())
					}
				}
			}
			typeParams := typeParamsBuilder.String()
			typeExprBuilder := &strings.Builder{}
			typeExprBuilder.WriteString(typeSpec.Name.String())
			if typeSpec.TypeParams != nil {
				typeExprBuilder.WriteString("<")
				for i, typeParam := range typeSpec.TypeParams.List {
					for j, name := range typeParam.Names {
						if i > 0 || j > 0 {
							typeExprBuilder.WriteString(", ")
						}
						typeExprBuilder.WriteString(name.String())
					}
				}
				typeExprBuilder.WriteString(">")
			}
			typeExpr := typeExprBuilder.String()

			// `gx::FieldTag` specializations
			tagIndex := 0
			for _, field := range typ.Fields.List {
				if field.Type != nil {
					for _, fieldName := range field.Names {
						if fieldName.IsExported() {
							builder.WriteString("template<")
							builder.WriteString(typeParams)
							builder.WriteString(">\nstruct gx::FieldTag<")
							builder.WriteString(typeExpr)
							builder.WriteString(", ")
							builder.WriteString(strconv.Itoa(tagIndex))
							builder.WriteString("> {\n")
							builder.WriteString("  inline static constexpr gx::FieldAttribs attribs { .name = \"")
							builder.WriteString(lowerFirst(fieldName.String()))
							builder.WriteString("\" };\n};\n")
							tagIndex++
						}
					}
				}
			}

			// `forEachField`
			if typeParams != "" {
				builder.WriteString("template<")
				builder.WriteString(typeParams)
				builder.WriteString(">\n")
			}
			builder.WriteString("inline void forEachField(")
			builder.WriteString(typeExpr)
			builder.WriteString(" &val, auto &&func) {\n")
			tagIndex = 0
			for _, field := range typ.Fields.List {
				if field.Type != nil {
					for _, fieldName := range field.Names {
						if fieldName.IsExported() {
							builder.WriteString("  func(gx::FieldTag<")
							builder.WriteString(typeExpr)
							builder.WriteString(", ")
							builder.WriteString(strconv.Itoa(tagIndex))
							builder.WriteString(">(), val.")
							builder.WriteString(fieldName.String())
							builder.WriteString(");\n")
							tagIndex++
						}
					}
				}
			}
			builder.WriteString("}")
		case *ast.InterfaceType:
			// Empty -- only used as generic constraint during typecheck
		default:
			// Empty -- alias declaration is definition
		}
		result = builder.String()
		c.genTypeMetas[typeSpec] = result
		return result
	}
}

//
// Functions
//

var fieldTagMethodNameRe = regexp.MustCompile(`^(.*)_([^_]*)$`)

func (c *Compiler) genFuncDecl(decl *ast.FuncDecl) string {
	if result, ok := c.genFuncDecls[decl]; ok {
		return result
	} else {
		obj := c.types.Defs[decl.Name]
		sig := obj.Type().(*types.Signature)
		recv := sig.Recv()

		builder := &strings.Builder{}

		// Type parameters
		addTypeParams := func(typeParams *types.TypeParamList) {
			if typeParams != nil {
				builder.WriteString("template<")
				for i, nTypeParams := 0, typeParams.Len(); i < nTypeParams; i++ {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString("typename ")
					builder.WriteString(typeParams.At(i).Obj().Name())
				}
				builder.WriteString(">\n")
			}
		}
		var recvNamedType *types.Named
		if recv != nil {
			switch recvType := recv.Type().(type) {
			case *types.Named:
				recvNamedType = recvType
				addTypeParams(recvType.TypeParams())
			case *types.Pointer:
				switch elemType := recvType.Elem().(type) {
				case *types.Named:
					recvNamedType = elemType
					addTypeParams(elemType.TypeParams())
				}
			}
		}
		addTypeParams(sig.TypeParams())

		// Return type
		if rets := sig.Results(); rets.Len() > 1 {
			c.errorf(decl.Type.Results.Pos(), "multiple return values not supported")
		} else if rets.Len() == 1 {
			ret := rets.At(0)
			builder.WriteString(c.genTypeExpr(ret.Type(), ret.Pos()))
		} else {
			if obj.Pkg().Name() == "main" && decl.Name.String() == "main" && recv == nil {
				builder.WriteString("int ")
			} else {
				builder.WriteString("void ")
			}
		}

		// Field tag
		name := decl.Name.String()
		fieldTag := ""
		if recvNamedType != nil {
			if structType, ok := recvNamedType.Underlying().(*types.Struct); ok {
				if matches := fieldTagMethodNameRe.FindStringSubmatch(name); len(matches) == 3 {
					name = matches[1]
					fieldName := matches[2]
					numFields := structType.NumFields()
					matchingTagIndex := -1
					tagIndex := 0
					for fieldIndex := 0; fieldIndex < numFields; fieldIndex++ {
						if field := structType.Field(fieldIndex); field.Exported() && !field.Embedded() {
							if field.Name() == fieldName {
								matchingTagIndex = tagIndex
							}
							tagIndex++
						}
					}
					typeExpr := trimFinalSpace(c.genTypeExpr(recvNamedType, recv.Pos()))
					if matchingTagIndex == -1 {
						c.errorf(decl.Name.Pos(), "struct %s has no field named %s", typeExpr, fieldName)
					} else {
						fieldTagBuilder := &strings.Builder{}
						fieldTagBuilder.WriteString("gx::FieldTag<")
						fieldTagBuilder.WriteString(typeExpr)
						fieldTagBuilder.WriteString(", ")
						fieldTagBuilder.WriteString(strconv.Itoa(matchingTagIndex))
						fieldTagBuilder.WriteString(">")
						fieldTag = fieldTagBuilder.String()
					}
				}
			}
		}

		// Name
		builder.WriteString(name)

		// Parameters
		builder.WriteByte('(')
		addParam := func(param *types.Var) {
			typ := param.Type()
			switch typ.Underlying().(type) {
			case *types.Array:
				c.errorf(param.Pos(), "cannot pass array by value, use pointer to array *%s instead", typ)
			case *types.Slice:
				c.errorf(param.Pos(), "cannot pass slice by value, use pointer to slice *%s instead", typ)
			}
			if _, ok := typ.(*types.Signature); ok {
				builder.WriteString("auto &&")
			} else {
				builder.WriteString(c.genTypeExpr(typ, param.Pos()))
			}
			builder.WriteString(param.Name())
		}
		if recv != nil {
			if fieldTag != "" {
				builder.WriteString(fieldTag)
				builder.WriteString(", ")
			}
			addParam(recv)
		}
		for i, nParams := 0, sig.Params().Len(); i < nParams; i++ {
			if i > 0 || recv != nil {
				builder.WriteString(", ")
			}
			addParam(sig.Params().At(i))
		}
		builder.WriteByte(')')

		result = builder.String()
		c.genFuncDecls[decl] = result
		return result
	}
}

//
// Expressions
//

func (c *Compiler) writeIdent(ident *ast.Ident) {
	typ := c.types.Types[ident]
	if typ.IsNil() {
		c.write("nullptr")
		return
	}
	if typ.IsBuiltin() {
		c.write("gx::")
	}
	if ext, ok := c.externs[c.types.Uses[ident]]; ok {
		c.write(ext)
	} else {
		c.write(ident.Name) // TODO: Package namespace
	}
}

func (c *Compiler) writeBasicLit(lit *ast.BasicLit) {
	switch lit.Kind {
	case token.INT:
		c.write(lit.Value)
	case token.FLOAT:
		c.write(lit.Value)
		c.write("f")
	case token.STRING:
		c.write(lit.Value)
	case token.CHAR:
		c.write(lit.Value)
	default:
		c.errorf(lit.Pos(), "unsupported literal kind")
	}
}

func (c *Compiler) writeFuncLit(lit *ast.FuncLit) {
	sig := c.types.TypeOf(lit).(*types.Signature)
	c.write("[&](")
	for i, nParams := 0, sig.Params().Len(); i < nParams; i++ {
		if i > 0 {
			c.write(", ")
		}
		param := sig.Params().At(i)
		c.write(c.genTypeExpr(param.Type(), param.Pos()))
		c.write(param.Name())
	}
	c.write(") ")
	c.writeBlockStmt(lit.Body)
	c.atBlockEnd = false
}

func (c *Compiler) writeCompositeLit(lit *ast.CompositeLit) {
	c.write(c.genTypeExpr(c.types.TypeOf(lit), lit.Pos()))
	c.write("{")
	if len(lit.Elts) > 0 {
		if _, ok := lit.Elts[0].(*ast.KeyValueExpr); ok {
			if typ, ok := c.types.TypeOf(lit.Type).Underlying().(*types.Struct); ok {
				c.computeFieldIndices(typ)
				lastIndex := 0
				for _, elt := range lit.Elts {
					field := c.types.ObjectOf(elt.(*ast.KeyValueExpr).Key.(*ast.Ident)).(*types.Var)
					if index := c.fieldIndices[field]; index < lastIndex {
						c.errorf(lit.Pos(), "struct literal fields must appear in definition order")
						break
					} else {
						lastIndex = index
					}
				}
			}
		}
		if c.fileSet.Position(lit.Pos()).Line == c.fileSet.Position(lit.Elts[0].Pos()).Line {
			c.write(" ")
			for i, elt := range lit.Elts {
				if i > 0 {
					c.write(", ")
				}
				c.writeExpr(elt)
			}
			c.write(" ")
		} else {
			c.write("\n")
			c.indent++
			for _, elt := range lit.Elts {
				c.writeExpr(elt)
				c.write(",\n")
			}
			c.indent--
		}
	}
	c.write("}")
}

func (c *Compiler) writeParenExpr(bin *ast.ParenExpr) {
	c.write("(")
	c.writeExpr(bin.X)
	c.write(")")
}

func (c *Compiler) writeSelectorExpr(sel *ast.SelectorExpr) {
	if basic, ok := c.types.TypeOf(sel.X).(*types.Basic); !(ok && basic.Kind() == types.Invalid) {
		c.writeExpr(sel.X)
		if _, ok := c.types.TypeOf(sel.X).(*types.Pointer); ok {
			c.write("->")
		} else {
			c.write(".")
		}
	}
	c.writeIdent(sel.Sel)
}

func (c *Compiler) writeIndexExpr(ind *ast.IndexExpr) {
	if _, ok := c.types.TypeOf(ind.X).(*types.Pointer); ok {
		c.write("(*(")
		c.writeExpr(ind.X)
		c.write("))")
	} else {
		c.writeExpr(ind.X)
	}
	c.write("[")
	c.writeExpr(ind.Index)
	c.write("]")
}

func (c *Compiler) writeCallExpr(call *ast.CallExpr) {
	method := false
	funType := c.types.Types[call.Fun]
	if _, ok := funType.Type.(*types.Signature); ok || funType.IsBuiltin() { // Function or method
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if sig, ok := c.types.Uses[sel.Sel].Type().(*types.Signature); ok && sig.Recv() != nil {
				method = true
				c.writeIdent(sel.Sel)
				c.write("(")
				_, xPtr := c.types.TypeOf(sel.X).(*types.Pointer)
				_, recvPtr := sig.Recv().Type().(*types.Pointer)
				if xPtr && !recvPtr {
					c.write("*(")
					c.writeExpr(sel.X)
					c.write(")")
				} else if !xPtr && recvPtr {
					c.write("&(")
					c.writeExpr(sel.X)
					c.write(")")
				} else {
					c.writeExpr(sel.X)
				}
			}
		}
		if !method {
			var typeArgs *types.TypeList
			switch fun := call.Fun.(type) {
			case *ast.Ident: // f(x)
				c.writeIdent(fun)
				typeArgs = c.types.Instances[fun].TypeArgs
			case *ast.SelectorExpr: // pkg.f(x)
				c.writeIdent(fun.Sel)
				typeArgs = c.types.Instances[fun.Sel].TypeArgs
			case *ast.IndexExpr:
				switch fun := fun.X.(type) {
				case *ast.Ident: // f[T](x)
					c.writeIdent(fun)
					typeArgs = c.types.Instances[fun].TypeArgs
				case *ast.SelectorExpr: // pkg.f[T](x)
					c.writeIdent(fun.Sel)
					typeArgs = c.types.Instances[fun.Sel].TypeArgs
				}
			default:
				c.writeExpr(fun)
			}
			if typeArgs != nil {
				c.write("<")
				for i, nTypeArgs := 0, typeArgs.Len(); i < nTypeArgs; i++ {
					if i > 0 {
						c.write(", ")
					}
					c.write(trimFinalSpace(c.genTypeExpr(typeArgs.At(i), call.Fun.Pos())))
				}
				c.write(">")
			}
			c.write("(")
		}
	} else { // Conversion
		typeExpr := trimFinalSpace(c.genTypeExpr(funType.Type, call.Fun.Pos()))
		if _, ok := call.Fun.(*ast.ParenExpr); ok {
			c.write("(")
			c.write(typeExpr)
			c.write(")")
		} else {
			c.write(typeExpr)
		}
		c.write("(")
	}
	for i, arg := range call.Args {
		if i > 0 || method {
			c.write(", ")
		}
		c.writeExpr(arg)
	}
	c.write(")")
}

func (c *Compiler) writeStarExpr(star *ast.StarExpr) {
	c.write("*")
	c.writeExpr(star.X)
}

func (c *Compiler) writeUnaryExpr(un *ast.UnaryExpr) {
	switch op := un.Op; op {
	case token.ADD, token.SUB, token.NOT:
		c.write(op.String())
	case token.AND:
		if !c.types.Types[un.X].Addressable() {
			c.errorf(un.OpPos, "cannot take address of a temporary object")
		}
		c.write(op.String())
	default:
		c.errorf(un.OpPos, "unsupported unary operator")
	}
	c.writeExpr(un.X)
}

func (c *Compiler) writeBinaryExpr(bin *ast.BinaryExpr) {
	c.writeExpr(bin.X)
	c.write(" ")
	switch op := bin.Op; op {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ,
		token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
		token.AND, token.OR, token.XOR, token.SHL, token.SHR,
		token.LAND, token.LOR:
		c.write(op.String())
	default:
		c.errorf(bin.OpPos, "unsupported binary operator")
	}
	c.write(" ")
	c.writeExpr(bin.Y)
}

func (c *Compiler) writeKeyValueExpr(kv *ast.KeyValueExpr) {
	if name, ok := kv.Key.(*ast.Ident); !ok {
		c.errorf(kv.Pos(), "unsupported literal key")
	} else {
		c.write(".")
		c.writeIdent(name)
		c.write(" = ")
		c.writeExpr(kv.Value)
	}
}

func (c *Compiler) writeExpr(expr ast.Expr) {
	switch expr := expr.(type) {
	case *ast.Ident:
		c.writeIdent(expr)
	case *ast.BasicLit:
		c.writeBasicLit(expr)
	case *ast.FuncLit:
		c.writeFuncLit(expr)
	case *ast.CompositeLit:
		c.writeCompositeLit(expr)
	case *ast.ParenExpr:
		c.writeParenExpr(expr)
	case *ast.SelectorExpr:
		c.writeSelectorExpr(expr)
	case *ast.IndexExpr:
		c.writeIndexExpr(expr)
	case *ast.CallExpr:
		c.writeCallExpr(expr)
	case *ast.StarExpr:
		c.writeStarExpr(expr)
	case *ast.UnaryExpr:
		c.writeUnaryExpr(expr)
	case *ast.BinaryExpr:
		c.writeBinaryExpr(expr)
	case *ast.KeyValueExpr:
		c.writeKeyValueExpr(expr)
	default:
		c.errorf(expr.Pos(), "unsupported expression type")
	}
}

//
// Statements
//

func (c *Compiler) writeExprStmt(exprStmt *ast.ExprStmt) {
	c.writeExpr(exprStmt.X)
}

func (c *Compiler) writeAssignStmt(assignStmt *ast.AssignStmt) {
	if len(assignStmt.Lhs) != 1 {
		c.errorf(assignStmt.Pos(), "multi-value assignment unsupported")
		return
	}
	if assignStmt.Tok == token.DEFINE {
		if typ, ok := c.types.TypeOf(assignStmt.Rhs[0]).(*types.Basic); ok && typ.Kind() == types.String {
			c.write("gx::String ")
		} else {
			c.write("auto ")
		}
	}
	c.writeExpr(assignStmt.Lhs[0])
	c.write(" ")
	switch op := assignStmt.Tok; op {
	case token.DEFINE:
		c.write("=")
	case token.ASSIGN,
		token.ADD_ASSIGN, token.SUB_ASSIGN, token.MUL_ASSIGN, token.QUO_ASSIGN, token.REM_ASSIGN,
		token.AND_ASSIGN, token.OR_ASSIGN, token.XOR_ASSIGN, token.SHL_ASSIGN, token.SHR_ASSIGN:
		c.write(op.String())
	default:
		c.errorf(assignStmt.TokPos, "unsupported assignment operator")
	}
	c.write(" ")
	c.writeExpr(assignStmt.Rhs[0])
}

func (c *Compiler) writeReturnStmt(retStmt *ast.ReturnStmt) {
	if len(retStmt.Results) > 1 {
		c.errorf(retStmt.Results[0].Pos(), "multiple return values not supported")
	} else if len(retStmt.Results) == 1 {
		c.write("return ")
		c.writeExpr(retStmt.Results[0])
	} else {
		c.write("return")
	}
}

func (c *Compiler) writeBranchStmt(branchStmt *ast.BranchStmt) {
	switch tok := branchStmt.Tok; tok {
	case token.BREAK, token.CONTINUE:
		c.write(tok.String())
	default:
		c.errorf(branchStmt.TokPos, "unsupported branch statement")
	}
}

func (c *Compiler) writeBlockStmt(block *ast.BlockStmt) {
	c.write("{\n")
	c.indent++
	c.writeStmtList(block.List)
	c.indent--
	c.write("}")
	c.atBlockEnd = true
}

func (c *Compiler) writeIfStmt(ifStmt *ast.IfStmt) {
	c.write("if (")
	if ifStmt.Init != nil {
		c.writeStmt(ifStmt.Init)
		c.write("; ")
	}
	c.writeExpr(ifStmt.Cond)
	c.write(") ")
	c.writeStmt(ifStmt.Body)
	if ifStmt.Else != nil {
		c.write(" else ")
		c.writeStmt(ifStmt.Else)
	}
}

func (c *Compiler) writeForStmt(forStmt *ast.ForStmt) {
	c.write("for (")
	if forStmt.Init != nil {
		c.writeStmt(forStmt.Init)
	}
	c.write("; ")
	if forStmt.Cond != nil {
		c.writeExpr(forStmt.Cond)
	}
	c.write("; ")
	if forStmt.Post != nil {
		c.writeStmt(forStmt.Post)
	}
	c.write(") ")
	c.writeStmt(forStmt.Body)
}

func (c *Compiler) writeRangeStmt(rangeStmt *ast.RangeStmt) {
	if rangeStmt.Tok == token.ASSIGN {
		c.errorf(rangeStmt.TokPos, "must use := in range statement")
	}
	var key *ast.Ident
	if rangeStmt.Key != nil {
		if ident, ok := rangeStmt.Key.(*ast.Ident); ok && ident.Name != "_" {
			key = ident
		}
	}
	c.write("for (")
	if key != nil {
		c.write("auto ")
		c.writeIdent(key)
		c.write(" = -1; ")
	}
	c.write("auto &")
	if rangeStmt.Value != nil {
		c.writeExpr(rangeStmt.Value)
	} else {
		if ident, ok := rangeStmt.Value.(*ast.Ident); ok && ident.Name != "_" {
			c.writeIdent(ident)
		} else {
			c.write("_ [[maybe_unused]]")
		}
	}
	c.write(" : ")
	c.writeExpr(rangeStmt.X)
	c.write(") {\n")
	c.indent++
	if key != nil {
		c.write("++")
		c.writeIdent(key)
		c.write(";\n")
	}
	c.writeStmtList(rangeStmt.Body.List)
	c.indent--
	c.write("}")
	c.atBlockEnd = true
}

func (c *Compiler) writeStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		c.writeExprStmt(stmt)
	case *ast.AssignStmt:
		c.writeAssignStmt(stmt)
	case *ast.ReturnStmt:
		c.writeReturnStmt(stmt)
	case *ast.BranchStmt:
		c.writeBranchStmt(stmt)
	case *ast.BlockStmt:
		c.writeBlockStmt(stmt)
	case *ast.IfStmt:
		c.writeIfStmt(stmt)
	case *ast.ForStmt:
		c.writeForStmt(stmt)
	case *ast.RangeStmt:
		c.writeRangeStmt(stmt)
	default:
		c.errorf(stmt.Pos(), "unsupported statement type")
	}
}

func (c *Compiler) writeStmtList(list []ast.Stmt) {
	for _, stmt := range list {
		c.writeStmt(stmt)
		if !c.atBlockEnd {
			c.write(";")
		}
		c.write("\n")
	}
}

//
// Top-level
//

//go:embed gx.hh
var preamble string

func (c *Compiler) compile() {
	// Initialize maps
	c.externs = make(map[types.Object]string)
	c.fieldIndices = make(map[*types.Var]int)
	c.genTypeExprs = make(map[types.Type]string)
	c.genTypeDecls = make(map[*ast.TypeSpec]string)
	c.genTypeDefns = make(map[*ast.TypeSpec]string)
	c.genTypeMetas = make(map[*ast.TypeSpec]string)
	c.genFuncDecls = make(map[*ast.FuncDecl]string)

	// Initialize builders
	c.errors = &strings.Builder{}
	c.outputCC = &strings.Builder{}
	c.outputHH = &strings.Builder{}

	// Load main package
	packagesConfig := &packages.Config{
		Mode: packages.NeedImports | packages.NeedDeps |
			packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}
	loadPkgs, err := packages.Load(packagesConfig, c.mainPkgPath)
	if err != nil {
		fmt.Fprintln(c.errors, err)
	}
	if len(loadPkgs) == 0 {
		return
	}
	for _, pkg := range loadPkgs {
		for _, err := range pkg.Errors {
			if err.Pos != "" {
				fmt.Fprintf(c.errors, "%s: %s\n", err.Pos, err.Msg)
			} else {
				fmt.Fprintln(c.errors, err.Msg)
			}
		}
	}
	if c.errored() {
		return
	}
	c.fileSet = loadPkgs[0].Fset

	// Collect packages
	var pkgs []*packages.Package
	{
		visited := make(map[*packages.Package]bool)
		var visit func(pkg *packages.Package)
		visit = func(pkg *packages.Package) {
			if !visited[pkg] {
				visited[pkg] = true
				for _, dep := range pkg.Imports {
					visit(dep)
				}
				pkgs = append(pkgs, pkg)
				if pkg.Fset != c.fileSet {
					c.errorf(0, "internal error: filesets differ")
				}
			}
		}
		for _, pkg := range loadPkgs {
			visit(pkg)
		}
	}

	// Collect type info
	c.types = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Instances:  make(map[*ast.Ident]types.Instance),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
	}
	for _, pkg := range pkgs {
		for k, v := range pkg.TypesInfo.Types {
			c.types.Types[k] = v
		}
		for k, v := range pkg.TypesInfo.Instances {
			c.types.Instances[k] = v
		}
		for k, v := range pkg.TypesInfo.Defs {
			c.types.Defs[k] = v
		}
		for k, v := range pkg.TypesInfo.Uses {
			c.types.Uses[k] = v
		}
		for k, v := range pkg.TypesInfo.Implicits {
			c.types.Implicits[k] = v
		}
		for k, v := range pkg.TypesInfo.Selections {
			c.types.Selections[k] = v
		}
		for k, v := range pkg.TypesInfo.Scopes {
			c.types.Scopes[k] = v
		}
	}

	// Collect top-level decls and externs
	var typeSpecs []*ast.TypeSpec
	exports := make(map[types.Object]bool)
	behaviors := make(map[types.Object]bool)
	var valueSpecs []*ast.ValueSpec
	var funcDecls []*ast.FuncDecl
	{
		externsRe := regexp.MustCompile(`//gx:externs (.*)`)
		externRe := regexp.MustCompile(`//gx:extern (.*)`)
		parseDirective := func(re *regexp.Regexp, doc *ast.CommentGroup) string {
			if doc != nil {
				for _, comment := range doc.List {
					if matches := re.FindStringSubmatch(comment.Text); len(matches) > 1 {
						return matches[1]
					}
				}
			}
			return ""
		}
		typeSpecVisited := make(map[*ast.TypeSpec]bool)
		valueSpecVisited := make(map[*ast.ValueSpec]bool)
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				fileExt := ""
				if len(file.Comments) > 0 {
					fileExt = parseDirective(externsRe, file.Comments[0])
				}
				for _, decl := range file.Decls {
					switch decl := decl.(type) {
					case *ast.GenDecl:
						declExt := parseDirective(externRe, decl.Doc)
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.TypeSpec:
								visitExternFields := func(typeSpec *ast.TypeSpec) {
									if typ, ok := typeSpec.Type.(*ast.StructType); ok {
										for _, field := range typ.Fields.List {
											for _, fieldName := range field.Names {
												c.externs[c.types.Defs[fieldName]] = lowerFirst(fieldName.String())
											}
										}
									}
								}
								var visitTypeSpec func(typeSpec *ast.TypeSpec, export bool)
								visitTypeSpec = func(typeSpec *ast.TypeSpec, export bool) {
									obj := c.types.Defs[typeSpec.Name]
									visited := typeSpecVisited[typeSpec]
									if visited && !(export && !exports[obj]) {
										return
									}
									if export {
										exports[obj] = true
									}
									if !visited {
										typeSpecVisited[typeSpec] = true
										if structType, ok := typeSpec.Type.(*ast.StructType); ok {
											for _, field := range structType.Fields.List {
												if field.Names == nil {
													if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "Behavior" {
														behaviors[obj] = true
														export = true
													}
												}
											}
										}
									}
									ast.Inspect(typeSpec, func(node ast.Node) bool {
										if ident, ok := node.(*ast.Ident); ok && ident.Obj != nil && ident.Obj.Decl != nil {
											if typeSpec, ok := ident.Obj.Decl.(*ast.TypeSpec); ok {
												visitTypeSpec(typeSpec, export)
											}
										}
										return true
									})
									if !visited {
										if specExt := parseDirective(externRe, typeSpec.Doc); specExt != "" {
											c.externs[c.types.Defs[typeSpec.Name]] = specExt
											visitExternFields(typeSpec)
										} else if declExt != "" {
											c.externs[c.types.Defs[typeSpec.Name]] = declExt
											visitExternFields(typeSpec)
										} else if fileExt != "" {
											c.externs[c.types.Defs[typeSpec.Name]] = fileExt + typeSpec.Name.String()
											visitExternFields(typeSpec)
										} else {
											typeSpecs = append(typeSpecs, typeSpec)
										}
									}
								}
								visitTypeSpec(spec, false)
							case *ast.ValueSpec:
								var visitValueSpec func(valueSpec *ast.ValueSpec)
								visitValueSpec = func(valueSpec *ast.ValueSpec) {
									if valueSpecVisited[valueSpec] {
										return
									}
									valueSpecVisited[valueSpec] = true
									ast.Inspect(valueSpec, func(node ast.Node) bool {
										if ident, ok := node.(*ast.Ident); ok && ident.Obj != nil && ident.Obj.Decl != nil {
											if valueSpec, ok := ident.Obj.Decl.(*ast.ValueSpec); ok {
												visitValueSpec(valueSpec)
											}
										}
										return true
									})
									specExt := parseDirective(externRe, spec.Doc)
									for _, name := range spec.Names {
										if specExt != "" {
											c.externs[c.types.Defs[name]] = specExt
										} else if declExt != "" {
											c.externs[c.types.Defs[name]] = declExt
										} else if fileExt != "" {
											c.externs[c.types.Defs[name]] = fileExt + name.String()
										}
									}
									if specExt == "" && declExt == "" && fileExt == "" {
										valueSpecs = append(valueSpecs, valueSpec)
									}
								}
								visitValueSpec(spec)
							}
						}
					case *ast.FuncDecl:
						if declExt := parseDirective(externRe, decl.Doc); declExt != "" {
							c.externs[c.types.Defs[decl.Name]] = declExt
						} else if fileExt != "" {
							c.externs[c.types.Defs[decl.Name]] = fileExt + decl.Name.String()
						} else {
							funcDecls = append(funcDecls, decl)
						}
					}
				}
			}
		}
	}

	// `#include`s
	var includes string
	{
		re := regexp.MustCompile(`//gx:include (.*)`)
		visited := make(map[string]bool)
		builder := &strings.Builder{}
		for _, pkg := range pkgs {
			for _, file := range pkg.Syntax {
				if len(file.Comments) > 0 {
					for _, comment := range file.Comments[0].List {
						if matches := re.FindStringSubmatch(comment.Text); len(matches) > 1 {
							include := matches[1]
							if !visited[include] {
								visited[include] = true
								builder.WriteString("#include ")
								builder.WriteString(include)
								builder.WriteString("\n")
							}
						}
					}
				}
			}
		}
		includes = builder.String()
	}

	// Output '.cc'
	{
		// Includes, preamble
		c.write(includes)
		c.write(preamble)

		// Types
		c.write("\n\n")
		c.write("//\n// Types\n//\n\n")
		for _, typeSpec := range typeSpecs {
			if typeDecl := c.genTypeDecl(typeSpec); typeDecl != "" {
				c.write(typeDecl)
				c.write(";\n")
			}
		}
		for _, typeSpec := range typeSpecs {
			if typeDefn := c.genTypeDefn(typeSpec); typeDefn != "" {
				c.write("\n")
				c.write(typeDefn)
				c.write(";\n")
			}
		}

		// Meta
		c.write("\n\n")
		c.write("//\n// Meta\n//\n")
		for _, typeSpec := range typeSpecs {
			if typeDecl := c.genTypeDecl(typeSpec); typeDecl != "" {
				if meta := c.genTypeMeta(typeSpec); meta != "" {
					c.write("\n")
					c.write(meta)
					c.write("\n")
				}
			}
		}

		// Function declarations
		c.write("\n\n")
		c.write("//\n// Function declarations\n//\n\n")
		for _, funcDecl := range funcDecls {
			c.write(c.genFuncDecl(funcDecl))
			c.write(";\n")
		}

		// Variables
		c.write("\n\n")
		c.write("//\n// Variables\n//\n\n")
		for _, valueSpec := range valueSpecs {
			for i, name := range valueSpec.Names {
				c.write("inline ")
				if name.Obj.Kind == ast.Con {
					c.write("constexpr ")
				}
				if valueSpec.Type != nil {
					c.write(c.genTypeExpr(c.types.TypeOf(valueSpec.Type), valueSpec.Type.Pos()))
				} else {
					c.write("auto ")
				}
				c.writeIdent(name)
				if len(valueSpec.Values) > 0 {
					c.write(" = ")
					c.writeExpr(valueSpec.Values[i])
				}
				c.write(";\n")
			}
		}

		// Function definitions
		c.write("\n\n")
		c.write("//\n// Function definitions\n//\n")
		for _, funcDecl := range funcDecls {
			if funcDecl.Body != nil {
				c.write("\n")
				c.write(c.genFuncDecl(funcDecl))
				c.write(" ")
				c.writeBlockStmt(funcDecl.Body)
				c.write("\n")
			}
		}
	}

	// Output '.hh'
	{
		// Includes, preamble
		c.outputHH.WriteString(includes)
		c.outputHH.WriteString(preamble)

		// Types
		c.outputHH.WriteString("\n\n")
		c.outputHH.WriteString("//\n// Types\n//\n\n")
		for _, typeSpec := range typeSpecs {
			if exports[c.types.Defs[typeSpec.Name]] {
				if typeDecl := c.genTypeDecl(typeSpec); typeDecl != "" {
					c.outputHH.WriteString(typeDecl)
					c.outputHH.WriteString(";\n")
				}
			}
		}
		for _, typeSpec := range typeSpecs {
			if exports[c.types.Defs[typeSpec.Name]] {
				if typeDefn := c.genTypeDefn(typeSpec); typeDefn != "" {
					c.outputHH.WriteString("\n")
					if behaviors[c.types.Defs[typeSpec.Name]] {
						c.outputHH.WriteString("ComponentTypeListAdd(")
						c.outputHH.WriteString(typeSpec.Name.String())
						c.outputHH.WriteString(");\n")
					}
					c.outputHH.WriteString(typeDefn)
					c.outputHH.WriteString(";\n")
				}
			}
		}

		// Meta
		c.outputHH.WriteString("\n\n")
		c.outputHH.WriteString("//\n// Meta\n//\n")
		for _, typeSpec := range typeSpecs {
			if exports[c.types.Defs[typeSpec.Name]] {
				if typeDecl := c.genTypeDecl(typeSpec); typeDecl != "" {
					if meta := c.genTypeMeta(typeSpec); meta != "" {
						c.outputHH.WriteString("\n")
						c.outputHH.WriteString(meta)
						c.outputHH.WriteString("\n")
					}
				}
			}
		}

		// Function declarations
		c.outputHH.WriteString("\n\n")
		c.outputHH.WriteString("//\n// Function declarations\n//\n\n")
		for _, funcDecl := range funcDecls {
			if funcDecl.Recv != nil {
				for _, recv := range funcDecl.Recv.List {
					export := false
					ast.Inspect(recv.Type, func(node ast.Node) bool {
						if export {
							return false
						}
						if ident, ok := node.(*ast.Ident); ok && exports[c.types.Uses[ident]] {
							export = true
							return false
						}
						return true
					})
					if export {
						c.outputHH.WriteString(c.genFuncDecl(funcDecl))
						c.outputHH.WriteString(";\n")
					}
				}
			}
		}
	}
}

//
// Main
//

func main() {
	// Arguments
	if len(os.Args) != 3 {
		fmt.Println("usage: gx <main_package_path> <output_prefix>")
		return
	}
	mainPkgPath := os.Args[1]
	outputPrefix := os.Args[2]

	// Compile
	c := Compiler{mainPkgPath: mainPkgPath}
	c.compile()

	// Print output
	if c.errored() {
		fmt.Println(c.errors)
		os.Exit(1)
	} else {
		readersEqual := func(a, b io.Reader) bool {
			bufA := make([]byte, 1024)
			bufB := make([]byte, 1024)
			for {
				nA, errA := io.ReadFull(a, bufA)
				nB, _ := io.ReadFull(b, bufB)
				if !bytes.Equal(bufA[:nA], bufB[:nB]) {
					return false
				}
				if errA == io.EOF {
					return true
				}
			}
		}
		writeFileIfChanged := func(path string, contents string) {
			byteContents := []byte(contents)
			if f, err := os.Open(path); err == nil {
				defer f.Close()
				if readersEqual(f, bytes.NewReader(byteContents)) {
					return
				}
			}
			ioutil.WriteFile(path, byteContents, 0644)
		}
		writeFileIfChanged(outputPrefix+".gx.cc", c.outputCC.String())
		writeFileIfChanged(outputPrefix+".gx.hh", c.outputHH.String())
	}
}
