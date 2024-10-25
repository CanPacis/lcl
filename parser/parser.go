package parser

import (
	"fmt"
	"iter"
	"slices"
	"strconv"
	"strings"

	"github.com/CanPacis/lcl/errs"
	"github.com/CanPacis/lcl/internal"
	"github.com/CanPacis/lcl/parser/ast"
	"github.com/CanPacis/lcl/parser/lexer"
	"github.com/CanPacis/lcl/parser/token"
)

type Parser struct {
	file  string
	lexer *lexer.Lexer

	current token.Token
	buffer  []token.Token

	errors []error
	ctx    *internal.Stack[Context]
}

func (p *Parser) advance() token.Token {
	current := p.current

	if len(p.buffer) > 0 {
		p.current = p.buffer[0]
		p.buffer = p.buffer[1:]
	} else {
		p.current = p.lexer.Next()
	}

	return current
}

func (p *Parser) skip() {
	for p.current.Kind == token.WHITESPACE || p.current.Kind == token.COMMENT {
		p.advance()
	}
}

func (p *Parser) error(err error) {
	if err != nil {
		p.errors = append(p.errors, err)
	}
}

func (p *Parser) expect(kind ...token.Kind) token.Token {
	if !slices.Contains(kind, p.current.Kind) {
		switch p.current.Kind {
		case token.UNTERM_STR, token.UNTERM_TEMP, token.UNTERM_TEMP_EXPR:
			p.error(&errs.SyntaxError{
				Err:   errs.ErrUnterminatedConstruct,
				Token: p.current,
			})
		default:
			var details string

			switch p.ctx.Last() {
			case SEQUENCE:
				details = "you might have forgotten a closing token"
			case STATEMENT:
				switch p.current.Kind {
				case token.DECLARE, token.IMPORT:
					details = fmt.Sprintf("%s is a top level statement, try moving it up", p.current.Kind.String())
				case token.IDENT:
					details = "only section statements and fn definitions are valid here"
				}
			}

			p.error(&errs.SyntaxError{
				Err:      errs.ErrUnexpectedToken,
				Token:    p.current,
				Expected: kind,
				Details:  details,
			})
		}
	}

	return p.advance()
}

func (p *Parser) seq(open, close token.Kind) iter.Seq[int] {
	return func(yield func(int) bool) {
		p.expect(open)
		p.skip()
		p.ctx.Push(SEQUENCE)

		i := 0
		for p.current.Kind != token.EOF && p.current.Kind != close {
			if !yield(i) {
				return
			}
			p.skip()
			i++
		}
		p.expect(close)
		p.ctx.Pop()
	}
}

func (p *Parser) Parse() (*ast.File, error) {
	p.ctx.Push(TOP_LEVEL)
	p.skip()

	fr := p.parseDeclStmt()
	p.skip()

	imports := []*ast.ImportStmt{}

	for p.current.Kind != token.EOF && p.current.Kind == token.IMPORT {
		imports = append(imports, p.parseImportStmt())
		p.skip()
	}

	stmts := []ast.Stmt{}
	for p.current.Kind != token.EOF {
		stmts = append(stmts, p.parseStmt())
		p.skip()
	}

	if len(p.errors) != 0 {
		return nil, errs.NewErrorSet(p.file, p.errors)
	}

	var end token.Position
	if len(stmts) > 0 {
		end = stmts[len(stmts)-1].End()
	} else {
		end = fr.End()
	}

	return &ast.File{
		Node:    ast.NewNode(ast.FileNode, fr.Start(), end),
		Decl:    fr,
		Imports: imports,
		Stmts:   stmts,
	}, nil
}

// Statements

func (p *Parser) parseStmt() ast.Stmt {
	p.ctx.Push(STATEMENT)
	defer p.ctx.Pop()

	switch p.current.Kind {
	case token.TYPE:
		return p.parseTypeDefStmt()
	case token.FN:
		return p.parseFnDefStmt()
	case token.SECTION:
		return p.parseSectionStmt()
	default:
		p.expect()
		return &ast.EmptyStmt{
			Stmt: ast.NewStmtNode(ast.EmptyStmtNode, p.current.Start, p.current.End),
		}
	}
}

func (p *Parser) parseDeclStmt() *ast.DeclStmt {
	start := p.expect(token.DECLARE)
	p.skip()
	name := p.parseIdentExpr()
	p.skip()

	targets := []*ast.DeclTarget{}
	for range p.seq(token.LEFT_PARENS, token.RIGHT_PARENS) {
		var name *ast.IdentExpr
		var tag *ast.StringLitExpr
		start := p.current

		switch p.current.Kind {
		case token.STRING:
			tag = p.parseStringExpr()
			p.skip()
			p.expect(token.AS)
			p.skip()
			name = p.parseIdentExpr()
		case token.IDENT:
			name = p.parseIdentExpr()
		default:
			p.expect(token.STRING, token.IDENT)
		}

		targets = append(targets, &ast.DeclTarget{
			Node: ast.NewNode(ast.DeclTargetNode, start.Start, p.current.End),
			Tag:  tag,
			Name: name,
		})
	}

	return &ast.DeclStmt{
		Stmt:    ast.NewStmtNode(ast.DeclStmtNode, start.Start, p.current.End),
		Name:    name,
		Targets: targets,
	}
}

func (p *Parser) parseImportStmt() *ast.ImportStmt {
	start := p.expect(token.IMPORT)
	p.skip()

	list := []*ast.IdentExpr{}

	if p.current.Kind != token.LEFT_PARENS {
		list = append(list, p.parseIdentExpr())

		return &ast.ImportStmt{
			Stmt: ast.NewStmtNode(ast.ImportStmtNode, start.Start, p.current.End),
			List: list,
		}
	}

	for range p.seq(token.LEFT_PARENS, token.RIGHT_PARENS) {
		list = append(list, p.parseIdentExpr())
	}

	return &ast.ImportStmt{
		Stmt: ast.NewStmtNode(ast.ImportStmtNode, start.Start, p.current.End),
		List: list,
	}
}

func (p *Parser) parseTypeDefStmt() *ast.TypeDefStmt {
	start := p.expect(token.TYPE)
	p.skip()
	name := p.parseIdentExpr()
	p.skip()
	typ := p.parseTypeExpr()

	return &ast.TypeDefStmt{
		Stmt: ast.NewStmtNode(ast.TypeDefStmtNode, start.Start, typ.End()),
		Name: name,
		Type: typ,
	}
}

func (p *Parser) parseFnDefStmt() *ast.FnDefStmt {
	start := p.expect(token.FN)

	params := []*ast.TypePair{}
	for i := range p.seq(token.LEFT_PARENS, token.RIGHT_PARENS) {
		params = append(params, p.parseTypePair(i))
	}

	p.skip()
	name := p.parseIdentExpr()
	p.skip()
	body := p.parseExpr()

	return &ast.FnDefStmt{
		Stmt:   ast.NewStmtNode(ast.FnDefStmtNode, start.Start, body.End()),
		Params: params,
		Name:   name,
		Body:   body,
	}
}

func (p *Parser) parseSectionStmt() *ast.SectionStmt {
	start := p.expect(token.SECTION)
	p.skip()
	name := p.parseIdentExpr()
	p.skip()

	list := []ast.Entry{}

	for range p.seq(token.LEFT_CURLY_BRACE, token.RIGHT_CURLY_BRACE) {
		list = append(list, p.parseEntry())
	}

	return &ast.SectionStmt{
		Stmt: ast.NewStmtNode(ast.SectionStmtNode, start.Start, p.current.End),
		Name: name,
		Body: list,
	}
}

// Entries

func (p *Parser) parseEntry() ast.Entry {
	p.ctx.Push(ENTRY)
	defer p.ctx.Pop()

	if p.current.Kind == token.SECTION {
		return p.parseSectionStmt()
	}

	isTemplate := false
	isPartitioned := false
	name := p.parseIdentExpr()

	params := []*ast.TypePair{}
	if p.current.Kind == token.LEFT_PARENS {
		isTemplate = true

		for i := range p.seq(token.LEFT_PARENS, token.RIGHT_PARENS) {
			params = append(params, p.parseTypePair(i))
		}
	}

	if p.current.Kind == token.STAR {
		isPartitioned = true
		p.advance()
	}
	p.skip()

	fields := []*ast.Field{}
	for range p.seq(token.LEFT_CURLY_BRACE, token.RIGHT_CURLY_BRACE) {
		fields = append(fields, p.parseField())
	}

	if isTemplate {
		return &ast.TemplateEntry{
			Node:        ast.NewNode(ast.TemplateEntryNode, name.Start(), p.current.End),
			Partitioned: isPartitioned,
			Name:        name,
			Fields:      fields,
			Params:      params,
		}
	}

	return &ast.KeyEntry{
		Node:   ast.NewNode(ast.KeyEntryNode, name.Start(), p.current.End),
		Name:   name,
		Fields: fields,
	}
}

func (p *Parser) parseField() *ast.Field {
	tag := p.parseIdentExpr()
	p.skip()

	var value ast.Expr
	switch p.current.Kind {
	case token.STRING:
		value = p.parseStringExpr()
	case token.TEMPLATE:
		value = p.parseTemplateExpr()
	default:
		p.expect(token.STRING, token.TEMPLATE)
	}

	return &ast.Field{
		Node:  ast.NewNode(ast.FieldNode, tag.Start(), value.End()),
		Tag:   tag,
		Value: value,
	}
}

// Expressions

func (p *Parser) parseExpr() ast.Expr {
	p.ctx.Push(EXPRESSION)
	defer p.ctx.Pop()

	var expr ast.Expr = p.parseBinaryExpr()
	p.skip()

	for p.current.Kind == token.QUESTION_MARK {
		p.advance()
		p.skip()
		lhs := p.parseExpr()
		p.skip()
		p.expect(token.COLON)
		p.skip()
		rhs := p.parseExpr()

		expr = &ast.TernaryExpr{
			Node:      ast.NewNode(ast.TernaryExprNode, expr.Start(), rhs.End()),
			Predicate: expr,
			Left:      lhs,
			Right:     rhs,
		}
	}

	return expr
}

var operators = []token.Kind{
	token.AND, token.OR,
	token.EQUALS, token.NOT_EQUALS,
	token.GT, token.GTE,
	token.LT, token.LTE,
}

func (p *Parser) parseBinaryExpr() ast.Expr {
	var expr ast.Expr = p.parseArithmeticExpr()
	p.skip()

	for slices.Contains(operators, p.current.Kind) {
		operator := p.advance()
		p.skip()
		rhs := p.parseArithmeticExpr()
		expr = &ast.BinaryExpr{
			Node:     ast.NewNode(ast.BinaryExprNode, expr.Start(), rhs.End()),
			Operator: operator,
			Left:     expr,
			Right:    rhs,
		}
	}

	return expr
}

func (p *Parser) parseArithmeticExpr() ast.Expr {
	var expr ast.Expr = p.parseArithmeticExpr2()
	p.skip()

	for p.current.Kind == token.PLUS || p.current.Kind == token.MINUS {
		operator := p.advance()
		p.skip()
		rhs := p.parseArithmeticExpr2()
		expr = &ast.ArithmeticExpr{
			Node:     ast.NewNode(ast.ArithmeticExprNode, expr.Start(), rhs.End()),
			Operator: operator,
			Left:     expr,
			Right:    rhs,
		}
	}

	return expr
}

func (p *Parser) parseArithmeticExpr2() ast.Expr {
	var expr ast.Expr = p.parseExponentExpr()
	p.skip()

	for p.current.Kind == token.STAR || p.current.Kind == token.FORWARD_SLASH {
		operator := p.advance()
		p.skip()
		rhs := p.parseExponentExpr()
		expr = &ast.ArithmeticExpr{
			Node:     ast.NewNode(ast.ArithmeticExprNode, expr.Start(), rhs.End()),
			Operator: operator,
			Left:     expr,
			Right:    rhs,
		}
	}

	return expr
}

func (p *Parser) parseExponentExpr() ast.Expr {
	var expr ast.Expr = p.parseIndexExpr()
	p.skip()

	for p.current.Kind == token.CARET {
		operator := p.advance()
		p.skip()
		rhs := p.parseExponentExpr()
		expr = &ast.ArithmeticExpr{
			Node:     ast.NewNode(ast.ArithmeticExprNode, expr.Start(), rhs.End()),
			Operator: operator,
			Left:     expr,
			Right:    rhs,
		}
	}

	return expr
}

func (p *Parser) parseIndexExpr() ast.Expr {
	var expr ast.Expr

	switch p.current.Kind {
	case token.IDENT:
		expr = p.parseCallExpr()
	case token.STRING, token.TEMPLATE, token.NUMBER, token.DOT:
		expr = p.parseBasicExpr()
	case token.LEFT_PARENS:
		expr = p.parseGroupExpr()
	}

	p.skip()
	for p.current.Kind == token.LEFT_SQUARE_BRACKET {
		p.advance()
		p.skip()
		index := p.parseExpr()
		p.skip()
		end := p.expect(token.RIGHT_SQUARE_BRACKET)

		expr = &ast.IndexExpr{
			Node:  ast.NewNode(ast.IndexExprNode, expr.Start(), end.End),
			Host:  expr,
			Index: index,
		}
	}

	return expr
}

func (p *Parser) parseCallExpr() ast.Expr {
	var expr ast.Expr = p.parseMemberExpr()

	for p.current.Kind == token.LEFT_PARENS {
		args := []ast.Expr{}

		for range p.seq(token.LEFT_PARENS, token.RIGHT_PARENS) {
			args = append(args, p.parseExpr())
		}

		expr = &ast.CallExpr{
			Node: ast.NewNode(ast.CallExprNode, expr.Start(), p.current.End),
			Fn:   expr,
			Args: args,
		}
	}

	return expr
}

func (p *Parser) parseMemberExpr() ast.Expr {
	var expr ast.Expr = p.parseImportExpr()

	for p.current.Kind == token.DOT {
		p.advance()
		rhs := p.parseIdentExpr()
		expr = &ast.MemberExpr{
			Node:  ast.NewNode(ast.MemberExprNode, expr.Start(), rhs.End()),
			Left:  expr,
			Right: rhs,
		}
	}

	return expr
}

func (p *Parser) parseImportExpr() ast.Expr {
	expr := p.parseIdentExpr()

	if p.current.Kind != token.DOUBLE_COLON {
		return expr
	}

	p.advance()
	rhs := p.parseIdentExpr()

	return &ast.ImportExpr{
		Node:  ast.NewNode(ast.ImportExprNode, expr.Start(), rhs.End()),
		Left:  expr,
		Right: rhs,
	}
}

func (p *Parser) parseBasicExpr() ast.Expr {
	switch p.current.Kind {
	case token.STRING:
		return p.parseStringExpr()
	case token.TEMPLATE:
		return p.parseTemplateExpr()
	case token.NUMBER:
		return p.parseNumberExpr()
	default:
		p.expect(token.STRING, token.TEMPLATE, token.NUMBER, token.DOT)
		return &ast.EmptyExpr{
			Node: ast.NewNode(ast.EmptyExprNode, p.current.Start, p.current.End),
		}
	}
}

func (p *Parser) parseGroupExpr() *ast.GroupExpr {
	start := p.expect(token.LEFT_PARENS)
	p.skip()
	expr := p.parseExpr()
	p.skip()
	end := p.expect(token.RIGHT_PARENS)

	return &ast.GroupExpr{
		Node: ast.NewNode(ast.GroupExprNode, start.Start, end.End),
		Expr: expr,
	}
}

func (p *Parser) parseStringExpr() *ast.StringLitExpr {
	expr := p.expect(token.STRING)
	return &ast.StringLitExpr{
		Node:  ast.NewNode(ast.StringLitExprNode, expr.Start, expr.End),
		Value: expr.Literal,
	}
}

func (p *Parser) parseTemplateExpr() *ast.TemplateLitExpr {
	start := p.current
	tokens := lexer.LexTemplate(p.current)

	parts := []any{}

	for i, t := range tokens {
		switch t.Kind {
		case token.UNKNOWN:
			parts = append(parts, t.Raw)
		case token.UNTERM_TEMP_EXPR:
			parts = append(parts, i)
		}
	}

	exprs := []ast.Expr{}
	for i, part := range parts {
		switch part := part.(type) {
		case string:
			literal := part
			if i == 0 || i == len(parts)-1 {
				literal = strings.Trim(literal, "`")
			}

			exprs = append(exprs, &ast.StringLitExpr{
				Node:  ast.NewNode(ast.StringLitExprNode, start.Start, start.End),
				Value: literal,
			})
		case int:
			p.buffer = tokens[part:]
			p.advance()
			p.advance()
			p.skip()
			for p.current.Kind != token.EOF && p.current.Kind != token.RIGHT_CURLY_BRACE {
				exprs = append(exprs, p.parseExpr())
				p.skip()
			}
			p.expect(token.RIGHT_CURLY_BRACE)
		}
	}

	p.buffer = []token.Token{}
	p.advance()

	return &ast.TemplateLitExpr{
		Node:  ast.NewNode(ast.TemplateLitExprNode, start.Start, start.End),
		Value: exprs,
	}
}

func (p *Parser) parseNumberExpr() *ast.NumberLitExpr {
	expr := p.expect(token.NUMBER)
	value, err := strconv.ParseFloat(expr.Literal, 64)
	if err != nil {
		p.error(&errs.SyntaxError{
			Err:   errs.ErrMalformedNumber,
			Token: expr,
		})
	}

	return &ast.NumberLitExpr{
		Node:  ast.NewNode(ast.NumberLitExprNode, expr.Start, expr.End),
		Value: value,
	}
}

func (p *Parser) parseIdentExpr() *ast.IdentExpr {
	expr := p.expect(token.IDENT)
	return &ast.IdentExpr{
		Node:  ast.NewNode(ast.IdentExprNode, expr.Start, expr.End),
		Value: expr.Literal,
	}
}

// Type Expressions

func (p *Parser) parseTypeExpr() ast.TypeExpr {
	p.ctx.Push(TYPE_EXPRESSION)
	defer p.ctx.Pop()

	var expr ast.TypeExpr = p.parseStructExpr()

	for p.current.Kind == token.LEFT_SQUARE_BRACKET {
		p.advance()
		end := p.expect(token.RIGHT_SQUARE_BRACKET)

		expr = &ast.ListTypeExpr{
			Node: ast.NewNode(ast.ListTypeExprNode, expr.Start(), end.End),
			Type: expr,
		}
	}

	return expr
}

func (p *Parser) parseStructExpr() ast.TypeExpr {
	if p.current.Kind == token.IDENT {
		member := p.parseImportExpr()

		switch member := member.(type) {
		case *ast.IdentExpr:
			return member
		case *ast.ImportExpr:
			return member
		default:
			return &ast.EmptyExpr{}
		}
	}

	start := p.current

	list := []*ast.TypePair{}
	for i := range p.seq(token.LEFT_CURLY_BRACE, token.RIGHT_CURLY_BRACE) {
		list = append(list, p.parseTypePair(i))
	}

	return &ast.StructLitExpr{
		Node:   ast.NewNode(ast.StructLitExprNode, start.Start, p.current.End),
		Fields: list,
	}
}

func (p *Parser) parseTypePair(i int) *ast.TypePair {
	name := p.expect(token.IDENT)
	p.expect(token.COLON)
	p.skip()
	typ := p.parseTypeExpr()

	return &ast.TypePair{
		Node:  ast.NewNode(ast.TypePairNode, name.Start, typ.End()),
		Name:  ast.NewIdent(name),
		Index: i,
		Type:  typ,
	}
}

func New(file *File) *Parser {
	parser := &Parser{
		file:  file.Name,
		lexer: lexer.New(file.source),
		ctx:   internal.NewStack[Context](),
	}
	parser.advance()
	return parser
}

func Parse(file *File) (*ast.File, error) {
	return New(file).Parse()
}

func ParseStmt(file *File) (ast.Stmt, error) {
	p := New(file)
	node := p.parseStmt()
	p.expect(token.EOF)

	if len(p.errors) != 0 {
		return nil, errs.NewErrorSet(p.file, p.errors)
	}

	return node, nil
}

func ParseExpr(file *File) (ast.Expr, error) {
	p := New(file)
	node := p.parseExpr()
	p.expect(token.EOF)

	if len(p.errors) != 0 {
		return nil, errs.NewErrorSet(p.file, p.errors)
	}

	return node, nil
}

func ParseTypeExpr(file *File) (ast.TypeExpr, error) {
	p := New(file)
	node := p.parseTypeExpr()
	p.expect(token.EOF)

	if len(p.errors) != 0 {
		return nil, errs.NewErrorSet(p.file, p.errors)
	}

	return node, nil
}
