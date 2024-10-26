package gogen

import (
	"fmt"
	goast "go/ast"
	"go/printer"
	gotoken "go/token"
	"strings"

	"github.com/CanPacis/lcl/ir"
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/parser/token"
	"github.com/CanPacis/lcl/types"
)

func OpToken(t token.Kind) gotoken.Token {
	switch t {
	case token.PLUS:
		return gotoken.ADD
	case token.MINUS:
		return gotoken.SUB
	case token.STAR:
		return gotoken.MUL
	case token.FORWARD_SLASH:
		return gotoken.QUO
	case token.AND:
		return gotoken.AND
	case token.OR:
		return gotoken.OR
	case token.LT:
		return gotoken.LSS
	case token.LTE:
		return gotoken.LEQ
	case token.GT:
		return gotoken.GTR
	case token.GTE:
		return gotoken.GEQ
	case token.EQUALS:
		return gotoken.EQL
	case token.NOT_EQUALS:
		return gotoken.NEQ
	default:
		return gotoken.ILLEGAL
	}
}

func ResolveExpr(expr ast.Expr) goast.Expr {
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		return &goast.BinaryExpr{
			X:  ResolveExpr(expr.Left),
			Y:  ResolveExpr(expr.Right),
			Op: OpToken(expr.Operator.Kind),
		}
	case *ast.ArithmeticExpr:
		return &goast.BinaryExpr{
			X:  ResolveExpr(expr.Left),
			Y:  ResolveExpr(expr.Right),
			Op: OpToken(expr.Operator.Kind),
		}
	case *ast.TernaryExpr:
		return &goast.CallExpr{
			Fun: goast.NewIdent("tern"),
			Args: []goast.Expr{
				ResolveExpr(expr.Predicate),
				ResolveExpr(expr.Left),
				ResolveExpr(expr.Right),
			},
		}
	case *ast.CallExpr:
		args := []goast.Expr{}

		for _, arg := range expr.Args {
			args = append(args, ResolveExpr(arg))
		}

		return &goast.CallExpr{
			Fun:  ResolveExpr(expr.Fn),
			Args: args,
		}
	case *ast.MemberExpr:
		return &goast.SelectorExpr{
			X:   ResolveExpr(expr.Left),
			Sel: goast.NewIdent(expr.Right.Value),
		}
	case *ast.ImportExpr:
		// TODO
		panic("not implemented")
	case *ast.IndexExpr:
		return &goast.IndexExpr{
			X:     ResolveExpr(expr.Host),
			Index: ResolveExpr(expr.Index),
		}
	case *ast.GroupExpr:
		return &goast.ParenExpr{X: ResolveExpr(expr.Expr)}
	case *ast.IdentExpr:
		return goast.NewIdent(expr.Value)
	case *ast.StringLitExpr:
		return &goast.BasicLit{Value: expr.Value, Kind: gotoken.STRING}
	case *ast.TemplateLitExpr:
		// TODO
		return &goast.BasicLit{Value: "TEMPLATE", Kind: gotoken.STRING}
	case *ast.NumberLitExpr:
		isInt := expr.Value == float64(int(expr.Value))
		if isInt {
			return &goast.BasicLit{Value: fmt.Sprintf("%d", int(expr.Value)), Kind: gotoken.INT}
		}
		return &goast.BasicLit{Value: fmt.Sprintf("%f", expr.Value), Kind: gotoken.FLOAT}
	case *ast.EmptyExpr:
		return nil
	default:
		return nil
	}
}

func ResolveTypeExpr(typ types.Type) goast.Expr {
	switch typ := typ.(type) {
	case *types.Constant:
		var name string
		switch typ.String() {
		case "bool":
			name = "bool"
		case "i8":
			name = "int8"
		case "i16":
			name = "int16"
		case "i32":
			name = "int32"
		case "i64":
			name = "int64"
		case "u8":
			name = "uint8"
		case "u16":
			name = "uint16"
		case "u32":
			name = "uint32"
		case "u64":
			name = "uint64"
		case "f32":
			name = "float32"
		case "f64":
			name = "float64"
		}

		return goast.NewIdent(name)
	case *types.Extended, *types.ExtIndexer:
		var name string
		switch typ.String() {
		case "int":
			name = "int"
		case "uint":
			name = "uint"
		case "byte":
			name = "byte"
		case "rune":
			name = "rune"
		case "string":
			name = "string"
		default:
			// TODO: implement this
			panic("not implemented")
		}
		return goast.NewIdent(name)
	case *types.List:
		return &goast.ArrayType{
			Elt: ResolveTypeExpr(typ.Type),
		}
	case *types.Template:
		t := ResolveTypeExpr(&types.Fn{
			In:  typ.In,
			Out: types.String,
		})

		// TODO: figure out how to return a generic type instead of this
		b := &strings.Builder{}
		b.WriteString("Template[")
		printer.Fprint(b, gotoken.NewFileSet(), t)
		b.WriteString("]")
		return goast.NewIdent(b.String())
	case *types.Struct:
		fields := []*goast.Field{}

		for _, pair := range *typ {
			fields = append(fields, &goast.Field{
				Names: []*goast.Ident{goast.NewIdent(exported(pair.Name))},
				Type:  ResolveTypeExpr(pair.Type),
			})
		}

		return &goast.StructType{
			Fields: &goast.FieldList{
				List: fields,
			},
		}
	case *types.Fn:
		params := []*goast.Field{}

		for _, param := range typ.In {
			params = append(params, &goast.Field{Type: ResolveTypeExpr(param)})
		}

		return &goast.FuncType{
			Params: &goast.FieldList{
				List: params,
			},
			Results: &goast.FieldList{
				List: []*goast.Field{{
					Type: ResolveTypeExpr(typ.Out),
				}},
			},
		}
	}
	return nil
}

func GenerateFuncDecl(fn *ir.FnDef, reciever string) *goast.FuncDecl {
	params := []*goast.Field{}

	for i, param := range fn.Stmt.Params {
		typ := fn.Type.In[i]

		params = append(params, &goast.Field{
			Names: []*goast.Ident{goast.NewIdent(param.Name.Value)},
			Type:  ResolveTypeExpr(typ),
		})
	}

	var rv *goast.FieldList

	if len(reciever) > 0 {
		rv = &goast.FieldList{
			List: []*goast.Field{
				{
					Names: []*goast.Ident{goast.NewIdent(recv(reciever))},
					Type:  goast.NewIdent(reciever),
				},
			},
		}
	}

	return &goast.FuncDecl{
		Name: goast.NewIdent(fn.Stmt.Name.Value),
		Body: &goast.BlockStmt{
			List: []goast.Stmt{
				&goast.ReturnStmt{
					Results: []goast.Expr{
						ResolveExpr(fn.Stmt.Body),
					},
				},
			},
		},
		Recv: rv,
		Type: &goast.FuncType{
			Params: &goast.FieldList{
				List: params,
			},
			Results: &goast.FieldList{
				List: []*goast.Field{
					{Type: ResolveTypeExpr(fn.Type.Out)},
				},
			},
		},
	}
}

func GenerateTypeDefDecl(def *ir.TypeDef) *goast.GenDecl {
	name := lower(def.Name)
	if def.Exported {
		name = exported(name)
	}

	return &goast.GenDecl{
		Tok: gotoken.TYPE,
		Specs: []goast.Spec{
			&goast.TypeSpec{
				Name: goast.NewIdent(name),
				Type: ResolveTypeExpr(def.Type),
			},
		},
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func lower(s string) string {
	return strings.ToLower(s)
}

func exported(s string) string {
	return capitalize(lower(s))
}

func recv(s string) string {
	return lower(string(s[0]))
}
