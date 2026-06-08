package dsl

import (
	"fmt"

	"example.com/mud/world/entities"
)

type ReactionDef struct {
	Commands []string   `parser:"@Ident { ',' @Ident }"`
	Rules    []*RuleDef `parser:"'{' { @@ } '}'"`
}

type RuleDef struct {
	When *WhenBlock `parser:"[ 'when' @@ ]"`
	Then *ThenBlock `parser:"'then' @@"`
}

type WhenBlock struct {
	Conds []*ConditionDef `parser:"'{' { @@ } '}'"`
}

type ThenBlock struct {
	Actions []*ActionDef `parser:"'{' { @@ } '}'"`
}

func (def *ReactionDef) Build() ([]*entities.Rule, error) {
	rules := make([]*entities.Rule, 0, len(def.Rules))
	for _, r := range def.Rules {
		rule, err := r.Build()
		if err != nil {
			return nil, fmt.Errorf("could not build reaction for %s: %w", def.Commands[0], err)
		}

		rules = append(rules, rule)
	}
	return rules, nil
}

func (def *RuleDef) Build() (*entities.Rule, error) {
	when, err := def.When.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build rule: %w", err)
	}

	then, err := def.Then.Build()
	if err != nil {
		return nil, fmt.Errorf("could not build rule: %w", err)
	}

	return &entities.Rule{
		When: when,
		Then: then,
	}, nil
}

func (def *WhenBlock) Build() ([]entities.Condition, error) {
	if def == nil {
		return []entities.Condition{}, nil
	}

	ret := make([]entities.Condition, len(def.Conds))

	for i, cDef := range def.Conds {
		condition, err := cDef.Build()
		if err != nil {
			return nil, fmt.Errorf("build when: %w", err)
		}
		ret[i] = condition
	}

	return ret, nil
}

func (def *ThenBlock) Build() ([]entities.Action, error) {
	ret := make([]entities.Action, len(def.Actions))

	for i, aDef := range def.Actions {
		action, err := aDef.Build()

		if err != nil {
			return nil, fmt.Errorf("build action: %w", err)
		}

		ret[i] = action
	}

	return ret, nil
}
