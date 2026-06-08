package dsl

import (
	"fmt"
	"strings"

	"example.com/mud/world/entities"
	"example.com/mud/world/entities/conditions"
)

type ConditionDef struct {
	Or *OrChain `parser:"@@"`
}

type OrChain struct {
	First *CondAtom     `parser:"@@"`
	Rest  []*OrChainRhs `parser:"( 'or' @@ )*"`
}

type OrChainRhs struct {
	Next *CondAtom `parser:"@@"`
}

type CondAtom struct {
	Paren      *ConditionDef             `parser:"  '(' @@ ')'"`
	Not        *NotCondition             `parser:"| @@"`
	Expr       *ExprCondition            `parser:"| @@"`
	HasTag     *HasTagCondition          `parser:"| @@"`
	IsPresent  *IsPresentCondition       `parser:"| @@"`
	RolesEqual *EventRolesEqualCondition `parser:"| @@"`
	HasChild   *HasChildCondition        `parser:"| @@"`
	MsgHas     *MessageContains          `parser:"| @@"`
}

type NotCondition struct {
	Cond *ConditionDef `parser:"'not' @@"`
}

type ExprCondition struct {
	Expr *Expression `parser:"'expr' '{' @@ '}'"`
}

type HasTagCondition struct {
	Target string `parser:"@Ident"`
	Tag    string `parser:"'has' 'tag' @String"`
}

type IsPresentCondition struct {
	Role string `parser:"@Ident 'exists'"`
}

type EventRolesEqualCondition struct {
	Role1 string `parser:"@Ident"`
	Role2 string `parser:"'is' @Ident"`
}

type HasChildCondition struct {
	ChildRole  string `parser:"@Ident"`
	ParentRole string `parser:"'in' @Ident"`
	Component  string `parser:"'.' @Ident"`
}

type MessageContains struct {
	Message string `parser:"'message' 'contains' @String"`
}

func (def *ConditionDef) Build() (entities.Condition, error) {
	if def == nil || def.Or == nil {
		return nil, fmt.Errorf("condition in when is empty")
	}

	acc, err := def.Or.First.Build()
	if err != nil {
		return nil, err
	}

	for _, rhs := range def.Or.Rest {
		next, err := rhs.Next.Build()
		if err != nil {
			return nil, err
		}
		acc = &conditions.Or{
			Left:  acc,
			Right: next,
		}
	}

	return acc, nil
}

func (def *CondAtom) Build() (entities.Condition, error) {
	if def == nil {
		return nil, fmt.Errorf("empty condition atom")
	}

	switch {
	case def.Paren != nil:
		return def.Paren.Build()
	case def.Not != nil:
		return def.Not.Build()
	case def.Expr != nil:
		return def.Expr.Build()
	case def.HasTag != nil:
		return def.HasTag.Build()
	case def.IsPresent != nil:
		return def.IsPresent.Build()
	case def.RolesEqual != nil:
		return def.RolesEqual.Build()
	case def.HasChild != nil:
		return def.HasChild.Build()
	case def.MsgHas != nil:
		return def.MsgHas.Build()
	}

	return nil, fmt.Errorf("unrecognized condition def")
}

func (def *NotCondition) Build() (entities.Condition, error) {
	inner, err := def.Cond.Build()
	if err != nil {
		return nil, fmt.Errorf("not condition: %w", err)
	}
	return &conditions.Not{Cond: inner}, nil
}

func (def *ExprCondition) Build() (entities.Condition, error) {
	expression, err := def.Expr.Build()
	if err != nil {
		return nil, fmt.Errorf("condition expression: %w", err)
	}
	return &conditions.ExpressionTrue{Expression: expression}, nil
}

func (def *HasTagCondition) Build() (entities.Condition, error) {
	eventRole, err := entities.ParseEventRole(def.Target)
	if err != nil {
		return nil, fmt.Errorf("could not build has tag condition: %w", err)
	}
	return &conditions.HasTag{
		EventRole: eventRole,
		Tag:       def.Tag,
	}, nil
}

func (def *IsPresentCondition) Build() (entities.Condition, error) {
	eventRole, err := entities.ParseEventRole(def.Role)
	if err != nil {
		return nil, fmt.Errorf("could not build is-present condition: %w", err)
	}
	return &conditions.IsPresent{EventRole: eventRole}, nil
}

func (def *EventRolesEqualCondition) Build() (entities.Condition, error) {
	role1, err := entities.ParseEventRole(def.Role1)
	if err != nil {
		return nil, fmt.Errorf("event roles equal condition: %w", err)
	}
	role2, err := entities.ParseEventRole(def.Role2)
	if err != nil {
		return nil, fmt.Errorf("event roles equal condition: %w", err)
	}
	return &conditions.EventRolesEqual{
		EventRole1: role1,
		EventRole2: role2,
	}, nil
}

func (def *HasChildCondition) Build() (entities.Condition, error) {
	parentRole, err := entities.ParseEventRole(def.ParentRole)
	if err != nil {
		return nil, fmt.Errorf("has child condition: %w", err)
	}
	component, err := entities.ParseComponentType(def.Component)
	if err != nil {
		return nil, fmt.Errorf("has child condition: %w", err)
	}
	childRole, err := entities.ParseEventRole(def.ChildRole)
	if err != nil {
		return nil, fmt.Errorf("has child condition: %w", err)
	}
	return &conditions.HasChild{
		ParentRole:    parentRole,
		ComponentType: component,
		ChildRole:     childRole,
	}, nil
}

func (def *MessageContains) Build() (entities.Condition, error) {
	return &conditions.MessageContains{
		MessageRegex: strings.ToLower(def.Message),
	}, nil
}
