package models

import "strings"

// ParseSyntax parses a command syntax string like "attack {target} with {instrument}"
// into a slice of PatTokens (the same format used by the DSL compiler).
func ParseSyntax(syntax string) []PatToken {
	parts := strings.Fields(syntax)
	tokens := make([]PatToken, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			slot := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			tokens = append(tokens, PatToken{SlotName: slot})
		} else {
			tokens = append(tokens, PatToken{Literal: part})
		}
	}
	return tokens
}
