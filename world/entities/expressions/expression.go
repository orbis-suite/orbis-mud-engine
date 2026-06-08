package expressions

import (
	"fmt"
	"math/rand"

	"example.com/mud/models"
	"example.com/mud/world/entities"
)

type Field struct {
	Role entities.EventRole
	Name string
}

type Expression interface {
	Eval(*entities.Event) (models.Value, error)
}

type ExpressionConst struct{ V models.Value }

func (ec *ExpressionConst) Eval(*entities.Event) (models.Value, error) { return ec.V, nil }

type ExpressionField struct{ F Field }

func (ef *ExpressionField) Eval(ev *entities.Event) (models.Value, error) {
	var e *entities.Entity
	switch ef.F.Role {
	case entities.EventRoleSource:
		e = ev.Source
	case entities.EventRoleInstrument:
		e = ev.Instrument
	case entities.EventRoleTarget:
		e = ev.Target
	case entities.EventRoleMessage:
		// message is a string, not an entity
		return models.VStr(ev.Message), nil
	default:
		return models.Value{}, fmt.Errorf("invalid role '%s' for expression", ef.F.Role)
	}

	return e.GetField(ef.F.Name), nil
}

type ExpressionDice struct {
	Count int
	Sides int
}

func (ed *ExpressionDice) Eval(ev *entities.Event) (models.Value, error) {
	if ed.Count <= 0 {
		return models.Value{}, fmt.Errorf("dice count must be > 0 (got %d)", ed.Count)
	}

	if ed.Count <= 0 {
		return models.Value{}, fmt.Errorf("dice sides must be > 0 (got %d)", ed.Sides)
	}

	total := 0
	for i := 0; i < ed.Count; i++ {
		total += 1 + rand.Intn(ed.Sides)
	}
	return models.VInt(total), nil
}

type ExpressionUnary struct {
	Op  UnaryOp
	Sub Expression
}

func (n *ExpressionUnary) Eval(ev *entities.Event) (models.Value, error) {
	v, err := n.Sub.Eval(ev)
	if err != nil {
		return models.Value{}, err
	}
	switch n.Op {
	case UNot:
		if v.K != models.KindBool {
			return models.Value{}, fmt.Errorf("! expects bool")
		}
		return models.VBool(!v.B), nil
	case UNeg:
		if v.K != models.KindInt {
			return models.Value{}, fmt.Errorf("- expects int")
		}
		return models.VInt(-v.I), nil
	default:
		return models.Value{}, fmt.Errorf("bad unary op")
	}
}

type ExpressionBinary struct {
	Op    BinaryOp
	Left  Expression
	Right Expression
}

func (n *ExpressionBinary) Eval(ev *entities.Event) (models.Value, error) {
	l, err := n.Left.Eval(ev)
	if err != nil {
		return models.Value{}, err
	}
	r, err := n.Right.Eval(ev)
	if err != nil {
		return models.Value{}, err
	}

	switch n.Op {
	case OpEq:
		return models.VBool(equals(l, r)), nil
	case OpNe:
		return models.VBool(!equals(l, r)), nil
	case OpGt, OpGe, OpLt, OpLe:
		if l.K != models.KindInt || r.K != models.KindInt {
			return models.Value{}, fmt.Errorf("comparison expects ints")
		}
		switch n.Op {
		case OpGt:
			return models.VBool(l.I > r.I), nil
		case OpGe:
			return models.VBool(l.I >= r.I), nil
		case OpLt:
			return models.VBool(l.I < r.I), nil
		case OpLe:
			return models.VBool(l.I <= r.I), nil
		}
	case OpAdd:
		if l.K == models.KindInt && r.K == models.KindInt {
			return models.VInt(l.I + r.I), nil
		}
		if l.K == models.KindString && r.K == models.KindString {
			return models.VStr(l.S + r.S), nil
		}
		return models.Value{}, fmt.Errorf("+ expects int+int or string+string")
	case OpSub, OpMul, OpDiv, OpDice:
		if l.K != models.KindInt || r.K != models.KindInt {
			return models.Value{}, fmt.Errorf("arithmetic expects ints")
		}
		switch n.Op {
		case OpSub:
			return models.VInt(l.I - r.I), nil
		case OpMul:
			return models.VInt(l.I * r.I), nil
		case OpDiv:
			if r.I == 0 {
				return models.Value{}, fmt.Errorf("division by zero")
			}
			return models.VInt(l.I / r.I), nil
		case OpDice:
			// Validate count (left) > 0
			if l.I <= 0 {
				return models.Value{}, fmt.Errorf("dice count must be > 0 (got %d)", l.I)
			}

			// Validate sides (right) > 0
			if r.I <= 0 {
				return models.Value{}, fmt.Errorf("dice sides must be > 0 (got %d)", r.I)
			}

			total := 0
			for i := 0; i < l.I; i++ {
				total += 1 + rand.Intn(r.I)
			}
			return models.VInt(total), nil
		}
	}
	return models.Value{}, fmt.Errorf("bad binary op")
}

func equals(a, b models.Value) bool {
	if a.K != b.K {
		return false
	}
	switch a.K {
	case models.KindInt:
		return a.I == b.I
	case models.KindString:
		return a.S == b.S
	case models.KindBool:
		return a.B == b.B
	case models.KindNil:
		return true
	default:
		return false
	}
}
