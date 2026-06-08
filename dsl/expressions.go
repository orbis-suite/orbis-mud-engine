package dsl

import (
	"fmt"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/expressions"
)

// Expressions are adapted from a Participle example (https://github.com/alecthomas/participle/blob/master/_examples/expr2/main.go)

type Expression struct {
	Equality *Equality `parser:"  @@"`

	// pairs are only used for room exits and are not ever passed to the evaluation
	// TODO support maps as a value in the Orbis Definition language
	Pairs []KV `parser:"| '{' @@ { ',' @@ } '}'"`
}

type Equality struct {
	Comparison *Comparison `parser:"@@"`
	Op         string      `parser:"( @( '!=' | '==' )"`
	Next       *Equality   `parser:"  @@ )*"`
}

type Comparison struct {
	Addition *Addition   `parser:"@@"`
	Op       string      `parser:"( @( '>' | '>=' | '<' | '<=' )"`
	Next     *Comparison `parser:"  @@ )*"`
}

type Addition struct {
	Multiplication *Multiplication `parser:"@@"`
	Op             string          `parser:"( @( '-' | '+' )"`
	Next           *Addition       `parser:"  @@ )*"`
}

type Multiplication struct {
	Unary *Unary          `parser:"@@"`
	Op    string          `parser:"( @( '/' | '*' | '$d' )"`
	Next  *Multiplication `parser:"  @@ )*"`
}

type Unary struct {
	Op      string   `parser:"  ( @( '!' | '-' | '$d' )"`
	Unary   *Unary   `parser:"   @@ )"`
	Primary *Primary `parser:"| @@"`
}

type Primary struct {
	Number        *int        `parser:"  @Int"`
	String        *string     `parser:"| @String"`
	Bool          *string     `parser:"| @( 'true' | 'false' )"`
	Field         *Field      `parser:"| @@"`
	SubExpression *Expression `parser:"| '(' @@ ')' "`
	Nil           bool        `parser:"| @'nil'"`
	List          *List       `parser:"| @@"`
}

type List struct {
	Numbers []int    `parser:"  '[' @Int { ',' @Int } ']'"`
	Strings []string `parser:"| '[' @String { ',' @String } ']'"`
	Bools   []string `parser:"| '[' ( 'true' | 'false' ) { ',' ( 'true' | 'false' ) } ']'"`
}

type Field struct {
	Role string `parser:"@Ident"`
	Name string `parser:"( '.' @Ident )?"`
}

// In the transition away from the old literals to expressions
// maps are no longer supported generally. This is to keep maps
// working for room exits.
func (e *Expression) AsMap() map[string]string {
	if len(e.Pairs) == 0 {
		return nil
	}
	m := make(map[string]string, len(e.Pairs))
	for _, kv := range e.Pairs {
		m[kv.Key] = kv.Value
	}
	return m
}

func (e *Expression) Build() (expressions.Expression, error) {
	return e.Equality.Build()
}

func (e *Equality) Build() (expressions.Expression, error) {
	left, err := e.Comparison.Build()
	if err != nil {
		return nil, err
	}

	curr := e.Next
	op := e.Op
	for curr != nil {
		right, err := e.Next.Build()
		if err != nil {
			return nil, err
		}

		bin, err := mapEqOp(op, left, right)
		if err != nil {
			return nil, err
		}
		left = bin

		op = curr.Op
		curr = curr.Next
	}
	return foldConst(left), nil
}

func (c *Comparison) Build() (expressions.Expression, error) {
	left, err := c.Addition.Build()
	if err != nil {
		return nil, err
	}

	curr := c.Next
	op := c.Op
	for curr != nil {
		right, err := c.Next.Build()
		if err != nil {
			return nil, err
		}

		bin, err := mapCmpOp(op, left, right)
		if err != nil {
			return nil, err
		}
		left = bin

		op = curr.Op
		curr = curr.Next
	}
	return foldConst(left), nil
}

func (a *Addition) Build() (expressions.Expression, error) {
	left, err := a.Multiplication.Build()
	if err != nil {
		return nil, err
	}

	curr := a.Next
	op := a.Op
	for curr != nil {
		right, err := curr.Multiplication.Build()
		if err != nil {
			return nil, err
		}
		bin, err := mapAddOp(op, left, right)
		if err != nil {
			return nil, err
		}
		left = bin
		op = curr.Op
		curr = curr.Next
	}
	return foldConst(left), nil
}

func (m *Multiplication) Build() (expressions.Expression, error) {
	left, err := m.Unary.Build()
	if err != nil {
		return nil, err
	}
	curr := m.Next
	op := m.Op
	for curr != nil {
		right, err := curr.Unary.Build()
		if err != nil {
			return nil, err
		}
		bin, err := mapMulOp(op, left, right)
		if err != nil {
			return nil, err
		}
		left = bin
		op = curr.Op
		curr = curr.Next
	}
	return foldConst(left), nil
}

func (u *Unary) Build() (expressions.Expression, error) {
	if u.Unary != nil {
		sub, err := u.Unary.Build()
		if err != nil {
			return nil, err
		}
		op, err := mapUnaryOp(u.Op)
		if err != nil {
			return nil, err
		}

		if op == expressions.UDice {
			// "d 6" is syntactical sugar for "1 d 6"
			return &expressions.ExpressionBinary{
				Op:    expressions.OpDice,
				Left:  &expressions.ExpressionConst{V: models.VInt(1)},
				Right: sub,
			}, nil
		}

		return foldConst(&expressions.ExpressionUnary{Op: op, Sub: sub}), nil
	}
	return u.Primary.Build()
}

func (p *Primary) Build() (expressions.Expression, error) {
	switch {
	case p.Number != nil:
		return &expressions.ExpressionConst{V: models.VInt(*p.Number)}, nil
	case p.String != nil:
		s := *p.String
		if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		return &expressions.ExpressionConst{V: models.VStr(s)}, nil
	case p.Bool != nil:
		return &expressions.ExpressionConst{V: models.VBool(*p.Bool == "true")}, nil
	case p.Nil:
		return &expressions.ExpressionConst{V: models.VNil()}, nil
	case p.Field != nil:
		eventRole, err := entities.ParseEventRole(p.Field.Role)
		if err != nil {
			return nil, fmt.Errorf("could not build has tag condition: %w", err)
		}
		return &expressions.ExpressionField{
			F: expressions.Field{
				Role: eventRole,
				Name: p.Field.Name,
			},
		}, nil
	case p.SubExpression != nil:
		return p.SubExpression.Build()
	case p.List != nil:
		// Numbers
		if len(p.List.Numbers) > 0 {
			il := make([]int, len(p.List.Numbers))
			copy(il, p.List.Numbers)
			return &expressions.ExpressionConst{
				V: models.Value{K: models.KindIntList, IL: il},
			}, nil
		}

		// Strings (tokens include quotes; strip them)
		if len(p.List.Strings) > 0 {
			sl := make([]string, len(p.List.Strings))
			for i, s := range p.List.Strings {
				if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
					s = s[1 : len(s)-1]
				}
				sl[i] = s
			}
			return &expressions.ExpressionConst{
				V: models.Value{K: models.KindStringList, SL: sl},
			}, nil
		}

		// Bools (parsed as "true"/"false" tokens)
		if len(p.List.Bools) > 0 {
			bl := make([]bool, len(p.List.Bools))
			for i, b := range p.List.Bools {
				bl[i] = (b == "true")
			}
			return &expressions.ExpressionConst{
				V: models.Value{K: models.KindBoolList, BL: bl},
			}, nil
		}

		return nil, fmt.Errorf("empty list literal")
	default:
		return nil, fmt.Errorf("invalid primary")
	}
}

func mapEqOp(tok string, l, r expressions.Expression) (expressions.Expression, error) {
	switch tok {
	case "==":
		return &expressions.ExpressionBinary{Op: expressions.OpEq, Left: l, Right: r}, nil
	case "!=":
		return &expressions.ExpressionBinary{Op: expressions.OpNe, Left: l, Right: r}, nil
	}
	return nil, fmt.Errorf("bad equality op %q", tok)
}
func mapCmpOp(tok string, l, r expressions.Expression) (expressions.Expression, error) {
	switch tok {
	case ">":
		return &expressions.ExpressionBinary{Op: expressions.OpGt, Left: l, Right: r}, nil
	case ">=":
		return &expressions.ExpressionBinary{Op: expressions.OpGe, Left: l, Right: r}, nil
	case "<":
		return &expressions.ExpressionBinary{Op: expressions.OpLt, Left: l, Right: r}, nil
	case "<=":
		return &expressions.ExpressionBinary{Op: expressions.OpLe, Left: l, Right: r}, nil
	}
	return nil, fmt.Errorf("bad comparison op %q", tok)
}
func mapAddOp(tok string, l, r expressions.Expression) (expressions.Expression, error) {
	switch tok {
	case "+":
		return &expressions.ExpressionBinary{Op: expressions.OpAdd, Left: l, Right: r}, nil
	case "-":
		return &expressions.ExpressionBinary{Op: expressions.OpSub, Left: l, Right: r}, nil
	}
	return nil, fmt.Errorf("bad add op %q", tok)
}
func mapMulOp(tok string, l, r expressions.Expression) (expressions.Expression, error) {
	switch tok {
	case "*":
		return &expressions.ExpressionBinary{Op: expressions.OpMul, Left: l, Right: r}, nil
	case "/":
		return &expressions.ExpressionBinary{Op: expressions.OpDiv, Left: l, Right: r}, nil
	case "$d":
		return &expressions.ExpressionBinary{Op: expressions.OpDice, Left: l, Right: r}, nil
	}
	return nil, fmt.Errorf("bad mul op %q", tok)
}

func mapUnaryOp(tok string) (expressions.UnaryOp, error) {
	switch tok {
	case "!":
		return expressions.UNot, nil
	case "-":
		return expressions.UNeg, nil
	case "$d":
		return expressions.UDice, nil
	}
	return 0, fmt.Errorf("bad unary op %q", tok)
}

func foldConst(n expressions.Expression) expressions.Expression {
	switch t := n.(type) {
	case *expressions.ExpressionUnary:
		k := foldConst(t.Sub)
		if sub, ok := k.(*expressions.ExpressionConst); ok {
			v, err := (&expressions.ExpressionUnary{Op: t.Op, Sub: sub}).Eval(nil)
			if err == nil {
				return &expressions.ExpressionConst{V: v}
			}
		}
		t.Sub = k
		return t
	case *expressions.ExpressionBinary:
		l := foldConst(t.Left)
		r := foldConst(t.Right)
		if lc, ok := l.(*expressions.ExpressionConst); ok {
			if rc, ok := r.(*expressions.ExpressionConst); ok {
				folded := &expressions.ExpressionBinary{Op: t.Op, Left: lc, Right: rc}

				// don't fold d operator, it should be random every time.
				if t.Op == expressions.OpDice {
					return folded
				}

				v, err := (folded).Eval(nil)
				if err == nil {
					return &expressions.ExpressionConst{V: v}
				}
			}
		}
		t.Left, t.Right = l, r
		return t
	default:
		return n
	}
}

// build and immediately evaluate an expression with an empty event into a value
func immediateEvalExpression(ex *Expression) (models.Value, error) {
	expr, err := ex.Build()
	if err != nil {
		return models.VNil(), fmt.Errorf("building expression during compilation: %w", err)
	}

	value, err := expr.Eval(nil)
	if err != nil {
		return models.VNil(), fmt.Errorf("evaluating expression during compilation: %w", err)
	}

	return value, nil
}

func immediateEvalExpressionAs(ex *Expression, expectedKind models.Kind) (models.Value, error) {
	value, err := immediateEvalExpression(ex)
	if err != nil {
		return models.VNil(), err
	}

	if value.K != expectedKind {
		return models.VNil(), fmt.Errorf("value from expression is of wrong type")
	}

	return value, nil
}
