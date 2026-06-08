package dsl

import (
	"fmt"
	"strings"

	"example.com/mud/models"
)

type CommandDef struct {
	Name   string          `parser:"@Ident"`
	Blocks []*CommandBlock `parser:"'{' { @@ } '}'"`
}

type CommandBlock struct {
	Field                *FieldDef             `parser:"  @@"`
	CommandDefinitionDef *CommandDefinitionDef `parser:"| @@"`
}

type CommandDefinitionDef struct {
	Fields []*FieldDef `parser:"'pattern' '{' { @@ } '}'"`
}

func (def *CommandDef) Build() (*models.CommandDefinition, error) {
	cmd := &models.CommandDefinition{
		Name:     strings.ToLower(def.Name),
		Aliases:  []string{},
		Patterns: []models.CommandPattern{},
	}

	for _, b := range def.Blocks {
		if b.Field != nil {
			f := b.Field
			switch f.Key {
			case "aliases":
				value, err := immediateEvalExpressionAs(f.Value, models.KindStringList)
				if err != nil {
					return nil, fmt.Errorf("could not get value '%s' for command aliases: %w", f.Key, err)
				}
				cmd.Aliases = append(cmd.Aliases, value.SL...)
			default:
				return nil, fmt.Errorf("unknown field '%s' in command definition", f.Key)
			}
		} else if b.CommandDefinitionDef != nil {
			commandPattern, err := b.CommandDefinitionDef.Build()
			if err != nil {
				return nil, fmt.Errorf("could not build command pattern: %w", err)
			}

			cmd.Patterns = append(cmd.Patterns, *commandPattern)
		} else {
			return nil, fmt.Errorf("could not expand command definition block")
		}
	}

	return cmd, nil
}

func (def *CommandDefinitionDef) Build() (*models.CommandPattern, error) {
	var p = &models.CommandPattern{
		Tokens: []models.PatToken{},
	}

	for _, f := range def.Fields {
		value, err := immediateEvalExpressionAs(f.Value, models.KindString)
		if err != nil {
			return nil, fmt.Errorf("could not get value '%s' for command: %w", f.Key, err)
		}

		switch f.Key {
		case "syntax":
			p.Tokens = tokenizeCommandSyntax(value.S)
		case "noMatch":
			p.NoMatchMessage = value.S
		case "help":
			p.HelpMessage = value.S
		default:
			err := fmt.Errorf("CommandDefinitionDef Field not recognized: %s", f.Key)
			return nil, err
		}
	}
	return p, nil
}

func tokenizeCommandSyntax(s string) []models.PatToken {
	var tokens []models.PatToken
	parts := strings.Fields(s)

	for _, part := range parts[:len(parts)-1] {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			slot := strings.Trim(part, "{}")
			tokens = append(tokens, models.Slot(slot))
		} else {
			tokens = append(tokens, models.Lit(part))
		}
	}

	lastPart := parts[len(parts)-1]
	if strings.HasPrefix(lastPart, "{") && strings.HasSuffix(lastPart, "}") {
		slot := strings.Trim(lastPart, "{}")
		if strings.Contains(slot, "...") {
			tokens = append(tokens, models.SlotRest(strings.TrimSuffix(slot, "...")))
		} else {
			tokens = append(tokens, models.Slot(slot))
		}
	} else {
		tokens = append(tokens, models.Lit(lastPart))
	}

	return tokens
}
