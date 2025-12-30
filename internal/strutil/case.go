package strutil

import "strings"

// ToPascalCase converts a string to PascalCase.
// Handles snake_case, kebab-case, and space-separated strings.
func ToPascalCase(s string) string {
	// Split by underscore, dash, or space
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}

	return strings.Join(parts, "")
}

// ToCamelCase converts a string to camelCase.
// Handles snake_case, kebab-case, space-separated, PascalCase, and already-camelCase strings.
func ToCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// If already camelCase or PascalCase (no underscores/dashes/spaces), just ensure first char is lowercase
	if !strings.ContainsAny(s, "_- ") {
		return strings.ToLower(s[0:1]) + s[1:]
	}

	// Otherwise, convert from snake_case/kebab-case/space-separated to camelCase
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return pascal
	}
	return strings.ToLower(pascal[0:1]) + pascal[1:]
}
