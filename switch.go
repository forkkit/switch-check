package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"sort"
)

func getSwitchesFromFile(path string) map[string][]string {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		log.Panic(err)
	}

	found := map[string][]string{}

	ast.Inspect(node, func(n ast.Node) bool {
		if n, ok := n.(*ast.SwitchStmt); ok {
			pos := fset.Position(n.Pos()).String()

			hasDefault := false
			for _, stmt := range n.Body.List {
				if caseClauseValues, ok := stmt.(*ast.CaseClause); ok {
					if len(caseClauseValues.List) == 0 {
						hasDefault = true
					}
					for _, caseValueValue := range caseClauseValues.List {
						found[pos] = append(found[pos], fmt.Sprintf("%v", caseValueValue))
					}
				}
			}

			// If there is a default case we should not include this switch
			// statement.
			if hasDefault {
				delete(found, pos)
			}
		}

		return true
	})

	return found
}

func findMissingValues(allValues map[string]Value, values []string) []string {
	found := false
	for _, value := range values {
		_, found = allValues[value]
		if found {
			break
		}
	}

	// If none of the input values (which were the case expressions) match any
	// of the known enum values this entire switch statement can be ignored.
	if !found {
		return nil
	}

	// Otherwise we assume that all values should appear.
	var missing []string
	ty := allValues[values[0]]
	for name, value := range allValues {
		if value.Type == ty.Type {
			missing = append(missing, name)
		}
	}

next:
	for i, have := range missing {
		for _, want := range values {
			if have == want {
				missing = append(missing[:i], missing[i+1:]...)
				goto next
			}
		}
	}

	sort.Strings(missing)

	return missing
}
