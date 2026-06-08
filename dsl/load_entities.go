package dsl

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"example.com/mud/models"
	"example.com/mud/world/entities"
	participle "github.com/alecthomas/participle/v2"
)

func LoadEntitiesFromDirectory(directoryName string) (map[string]*entities.Entity, []*models.CommandDefinition, error) {
	parser, err := participle.Build[DSL](
		participle.Lexer(DslLexer),
		participle.Elide("Whitespace", "Comment"),
		participle.Unquote("String"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("parser build failed %w", err)
	}

	var ast = &DSL{}

	err = filepath.WalkDir(directoryName, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("something went wrong: %v", err)
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".mud") {
			return nil
		}

		fmt.Println("Parsing DSL file:", path)

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %v", path, err)
		}

		fileSyntaxTree, err := parser.ParseString("", string(data))
		if err != nil {
			return fmt.Errorf("failed to parse %s: %v", path, err)
		}

		ast.Declarations = append(ast.Declarations, fileSyntaxTree.Declarations...)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("error walking DSL directory: %w", err)
	}

	entities, commands, err := Compile(ast)
	return entities, commands, err
}
