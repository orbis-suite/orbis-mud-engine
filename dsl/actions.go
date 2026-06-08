package dsl

import (
	"fmt"
	"time"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	"example.com/mud/world/entities/actions"
)

type ActionDef struct {
	Print                   *PrintAction             `parser:"  'print' @@"`
	Publish                 *PublishAction           `parser:"| 'publish' @@"`
	Copy                    *CopyAction              `parser:"| 'copy' @@"`
	Move                    *MoveAction              `parser:"| 'move' @@"`
	SetField                *SetFieldAction          `parser:"| 'set' @@"`
	DestroyAction           *DestroyAction           `parser:"| 'destroy' @@"`
	ScheduleOnceAction      *ScheduleOnceAction      `parser:"| @@"`
	ScheduleRepeatingAction *ScheduleRepeatingAction `parser:"| @@"`
	RevealChildrenAction    *RevealChildrenAction    `parser:"| @@"`
	ConditionalAction       *ConditionalAction       `parser:"| @@"`
}

type PrintAction struct {
	Target string `parser:"@Ident"`
	Value  string `parser:"@String"`
}

type PublishAction struct {
	Value string `parser:"@String"`
}

type CopyAction struct {
	EntityId  string `parser:"@String"`
	Target    string `parser:"'to' @Ident"`
	Component string `parser:"'.' @Ident"`
}

type MoveAction struct {
	RoleObject      string `parser:"@Ident"`
	RoleDestination string `parser:"'to' @Ident"`
	Component       string `parser:"'.' @Ident"`
}

type SetFieldAction struct {
	Role  string     `parser:"@Ident"`
	Field string     `parser:"'.' @Ident"`
	Expr  Expression `parser:"'to' @@"`
}

type RevealChildrenAction struct {
	Set       string `parser:"@('reveal' | 'hide')"`
	Role      string `parser:"@Ident"`
	Component string `parser:"'.' @Ident"`
}

type ScheduleOnceAction struct {
	ExprIn *Expression `parser:"'in' @@"`
	Units  string      `parser:"@( 'second' | 'seconds' | 'minute' | 'minutes' )"`
	Then   *ThenBlock  `parser:"@@"`
}

type ScheduleRepeatingAction struct {
	ExprIn *Expression `parser:"'repeat' 'every' @@"`
	Units  string      `parser:"@( 'second' | 'seconds' | 'minute' | 'minutes' )"`
	While  *IfDef      `parser:"'while' @@"`
}

type ConditionalAction struct {
	If      *IfDef   `parser:"'if' @@"`
	ElseIfs []*IfDef `parser:"{ 'else' 'if' @@ }"`
	Else    *ElseDef `parser:"[ @@ ]"`
}

type IfDef struct {
	When *WhenBlock `parser:"@@"`
	Then *ThenBlock `parser:"'then' @@"`
}

type ElseDef struct {
	Then *ThenBlock `parser:"'else' @@"`
}

type DestroyAction struct {
	Role string `parser:"@Ident"`
}

func (def *ActionDef) Build() (entities.Action, error) {
	switch {
	case def.Print != nil:
		return def.Print.Build()
	case def.Publish != nil:
		return def.Publish.Build()
	case def.Copy != nil:
		return def.Copy.Build()
	case def.Move != nil:
		return def.Move.Build()
	case def.SetField != nil:
		return def.SetField.Build()
	case def.DestroyAction != nil:
		return def.DestroyAction.Build()
	case def.RevealChildrenAction != nil:
		return def.RevealChildrenAction.Build()
	case def.ConditionalAction != nil:
		return def.ConditionalAction.Build()
	case def.ScheduleOnceAction != nil:
		return def.ScheduleOnceAction.Build()
	case def.ScheduleRepeatingAction != nil:
		return def.ScheduleRepeatingAction.Build()
	}

	return nil, fmt.Errorf("action is empty")
}

func (def *PrintAction) Build() (entities.Action, error) {
	eventRole, err := entities.ParseEventRole(def.Target)
	if err != nil {
		return nil, fmt.Errorf("could not build print action: %w", err)
	}

	return &actions.Print{
		Text:      def.Value,
		EventRole: eventRole,
	}, nil
}

func (def *PublishAction) Build() (entities.Action, error) {
	return &actions.Publish{
		Text: def.Value,
	}, nil
}

func (def *CopyAction) Build() (entities.Action, error) {
	eventRole, err := entities.ParseEventRole(def.Target)
	if eventRole == entities.EventRoleUnknown {
		return nil, fmt.Errorf("could not build copy action: %w", err)
	}

	component, err := entities.ParseComponentType(def.Component)
	if err != nil {
		return nil, fmt.Errorf("could not build action: %w", err)
	}

	return &actions.Copy{
		EntityId:      def.EntityId,
		EventRole:     eventRole,
		ComponentType: component,
	}, nil
}

func (def *MoveAction) Build() (entities.Action, error) {
	roleObject, err := entities.ParseEventRole(def.RoleObject)
	if err != nil {
		return nil, fmt.Errorf("could not build move action for origin: %w", err)
	}

	roleDestination, err := entities.ParseEventRole(def.RoleDestination)
	if err != nil {
		return nil, fmt.Errorf("could not build move action for destination: %w", err)
	}

	component, err := entities.ParseComponentType(def.Component)
	if err != nil {
		return nil, fmt.Errorf("could not build action: %w", err)
	}

	return &actions.Move{
		RoleObject:      roleObject,
		RoleDestination: roleDestination,
		ComponentType:   component,
	}, nil
}

func (def *SetFieldAction) Build() (entities.Action, error) {
	role, err := entities.ParseEventRole(def.Role)
	if err != nil {
		return nil, fmt.Errorf("event set field action: %w", err)
	}

	expression, err := def.Expr.Build()
	if err != nil {
		return nil, fmt.Errorf("expression set field action: %w", err)
	}

	return &actions.SetField{
		Role:       role,
		Field:      def.Field,
		Expression: expression,
	}, nil
}

func (def *DestroyAction) Build() (entities.Action, error) {
	role, err := entities.ParseEventRole(def.Role)
	if err != nil {
		return nil, fmt.Errorf("event destroy action: %w", err)
	}

	return &actions.Destroy{
		Role: role,
	}, nil
}

func (def *RevealChildrenAction) Build() (entities.Action, error) {
	role, err := entities.ParseEventRole(def.Role)
	if err != nil {
		return nil, fmt.Errorf("could not build reveal children action for role: %w", err)
	}

	component, err := entities.ParseComponentType(def.Component)
	if err != nil {
		return nil, fmt.Errorf("could not build reveal children action: %w", err)
	}

	return &actions.RevealChildren{
		Role:          role,
		ComponentType: component,
		Reveal:        def.Set == "reveal",
	}, nil
}

func (def *ScheduleOnceAction) Build() (entities.Action, error) {
	value, err := immediateEvalExpressionAs(def.ExprIn, models.KindInt)
	if err != nil {
		return nil, fmt.Errorf("expression value in schedule once expected int: %w", err)
	}

	var unitMultiplier time.Duration
	switch def.Units {
	case "second", "seconds":
		unitMultiplier = time.Second
	case "minute", "minutes":
		unitMultiplier = time.Minute
	default:
		return nil, fmt.Errorf("invalid unit in schedule once action: %w", err)
	}

	then, err := def.Then.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build schedule once then actions: %w", err)
	}

	return &actions.ScheduleOnce{
		Nanoseconds: time.Duration(value.I) * unitMultiplier,
		Actions:     then,
	}, nil
}

func (def *ScheduleRepeatingAction) Build() (entities.Action, error) {
	value, err := immediateEvalExpressionAs(def.ExprIn, models.KindInt)
	if err != nil {
		return nil, fmt.Errorf("expression value in schedule repeating expected int: %w", err)
	}

	var unitMultiplier time.Duration
	switch def.Units {
	case "second", "seconds":
		unitMultiplier = time.Second
	case "minute", "minutes":
		unitMultiplier = time.Minute
	default:
		return nil, fmt.Errorf("invalid unit in schedule repeating action: %w", err)
	}

	ruleDef := RuleDef{
		When: def.While.When,
		Then: def.While.Then,
	}

	rule, err := ruleDef.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build rule for schedule repeating action: %w", err)
	}

	return &actions.ScheduleRepeating{
		Nanoseconds: time.Duration(value.I) * unitMultiplier,
		Rule:        rule,
	}, nil
}

func (def *ConditionalAction) Build() (entities.Action, error) {
	ruleChain := make([]*entities.Rule, 0, len(def.ElseIfs)+1)

	// add first if
	ruleDef := RuleDef{
		When: def.If.When,
		Then: def.If.Then,
	}

	rule, err := ruleDef.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build 'if' action: %w", err)
	}

	ruleChain = append(ruleChain, rule)

	// add the rest of the "else if"s to the chain
	for _, elseIf := range def.ElseIfs {
		ruleDef := RuleDef{
			When: elseIf.When,
			Then: elseIf.Then,
		}

		rule, err := ruleDef.Build()
		if err != nil {
			return nil, fmt.Errorf("could not build 'else if' action: %w", err)
		}

		ruleChain = append(ruleChain, rule)
	}

	// add the "else" actions
	if def.Else != nil {
		elseDef := RuleDef{
			Then: def.Else.Then,
		}

		elseRule, err := elseDef.Build()
		if err != nil {
			return nil, fmt.Errorf("could not build else actions: %w", err)
		}

		ruleChain = append(ruleChain, elseRule)

	}

	return &actions.Conditional{
		RuleChain: ruleChain,
	}, nil
}
