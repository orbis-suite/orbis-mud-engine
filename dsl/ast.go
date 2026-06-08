package dsl

import (
	"github.com/alecthomas/participle/v2/lexer"
)

var DslLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Comment", Pattern: `//[^\n]*|(?s)/\*.*?\*/`},
	{Name: "String", Pattern: "\"([^\\\"\\\\]|\\\\.)*\"|'([^'\\\\]|\\\\.)*'|`[^`]*`"}, // ", ', or ` strings
	{Name: "Bool", Pattern: `\b(?:true|false)\b`},
	{Name: "Int", Pattern: `[0-9]+`},

	{Name: "Op", Pattern: `==|!=|<=|>=|\|\||&&|\$d`},
	{Name: "Punct", Pattern: `[\[\](){}.,=<>+\-*/!]|:`},

	{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "AtIdent", Pattern: `@[a-zA-Z_][a-zA-Z0-9_]*`},
	{Name: "Tag", Pattern: `#[a-zA-Z_][a-zA-Z0-9_]*`},

	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
})

type DSL struct {
	Declarations []*TopLevel `parser:"@@*"`
}

type TopLevel struct {
	Entity  *EntityDef  `parser:"'entity' @@"`
	Trait   *TraitDef   `parser:"| 'trait' @@"`
	Command *CommandDef `parser:"| 'command' @@"`
}

type EntityDef struct {
	Name   string         `parser:"@Ident"`
	Blocks []*EntityBlock `parser:"'{' { @@ } '}'"`
}

type TraitDef struct {
	Name   string         `parser:"@Ident"`
	Blocks []*EntityBlock `parser:"'{' { @@ } '}'"`
}

type EntityBlock struct {
	Component *ComponentDef        `parser:"  'component' @@"`
	Trait     *TraitInheritanceDef `parser:"| 'trait' @@"`
	Reaction  *ReactionDef         `parser:"| 'react' @@"`
	Field     *FieldDef            `parser:"| @@"`
}

type TraitInheritanceDef struct {
	Name   string      `parser:"@Ident"`
	Fields []*FieldDef `parser:"( '{' { @@ } '}' )?"`
}

type FieldDef struct {
	Key   string      `parser:"@Ident 'is'"`
	Value *Expression `parser:"@@"`
	Pairs []KV        `parser:"| '{' @@ { ',' @@ } '}'"`
}

type KV struct {
	Key   string `parser:"@String"`
	Value string `parser:"':' @String"`
}
