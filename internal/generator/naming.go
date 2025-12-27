package generator

import (
	"github.com/iancoleman/strcase"
)

// NamingConverter handles name conversions using strcase library
type NamingConverter struct{}

// NewNamingConverter creates a new NamingConverter instance
func NewNamingConverter() *NamingConverter {
	return &NamingConverter{}
}

// ToPascalCaseStrcase converts a string to PascalCase using strcase library
func (nc *NamingConverter) ToPascalCaseStrcase(s string) string {
	return strcase.ToCamel(s)
}

// ToSnakeCaseStrcase converts a string to snake_case using strcase library
func (nc *NamingConverter) ToSnakeCaseStrcase(s string) string {
	return strcase.ToSnake(s)
}

// ToGoFieldName converts a column name to a Go field name (PascalCase with acronym handling)
func (nc *NamingConverter) ToGoFieldName(columnName string) string {
	// Use strcase for base conversion
	pascalCase := strcase.ToCamel(columnName)

	// Handle common acronyms that strcase might not handle correctly
	return handleAcronyms(pascalCase)
}

// ToGoStructName converts a table name to a Go struct name (singular PascalCase)
func (nc *NamingConverter) ToGoStructName(tableName string) string {
	// First singularize, then convert to PascalCase
	singular := singularize(tableName)
	return strcase.ToCamel(singular)
}

// ToFileName converts a table name to a file name (snake_case.go)
func (nc *NamingConverter) ToFileName(tableName string) string {
	return strcase.ToSnake(tableName) + ".go"
}

// handleAcronyms handles common acronyms in Go naming
func handleAcronyms(s string) string {
	// Common acronyms that should be all uppercase
	acronyms := map[string]string{
		"Id":   "ID",
		"Url":  "URL",
		"Api":  "API",
		"Http": "HTTP",
		"Json": "JSON",
		"Xml":  "XML",
		"Sql":  "SQL",
		"Uuid": "UUID",
		"Ip":   "IP",
		"Html": "HTML",
		"Css":  "CSS",
		"Db":   "DB",
	}

	result := s
	for pattern, replacement := range acronyms {
		// Only replace at word boundaries (start, after lowercase, or at end)
		result = replaceAcronym(result, pattern, replacement)
	}
	return result
}

// replaceAcronym replaces an acronym while preserving word boundaries
func replaceAcronym(s, pattern, replacement string) string {
	// Simple replacement for common patterns
	// This handles cases like "UserId" -> "UserID", "ApiUrl" -> "APIURL"
	result := s

	// Check if pattern exists in string
	for i := 0; i <= len(result)-len(pattern); i++ {
		if result[i:i+len(pattern)] == pattern {
			// Check if this is at a word boundary (end of string or followed by uppercase)
			isWordBoundary := i+len(pattern) >= len(result) ||
				(result[i+len(pattern)] >= 'A' && result[i+len(pattern)] <= 'Z')
			if isWordBoundary {
				result = result[:i] + replacement + result[i+len(pattern):]
				i += len(replacement) - 1
			}
		}
	}

	return result
}

// singularize converts a plural table name to singular
// This is a simple implementation; consider using a library like "github.com/jinzhu/inflection" for production
func singularize(word string) string {
	if word == "" {
		return word
	}

	// Common irregular plurals
	irregulars := map[string]string{
		"people":   "person",
		"children": "child",
		"men":      "man",
		"women":    "woman",
		"teeth":    "tooth",
		"feet":     "foot",
		"mice":     "mouse",
		"geese":    "goose",
	}

	if singular, ok := irregulars[word]; ok {
		return singular
	}

	// Handle common plural endings
	if len(word) > 3 {
		// -ies -> -y (e.g., categories -> category)
		if word[len(word)-3:] == "ies" {
			return word[:len(word)-3] + "y"
		}

		// -ves -> -f (e.g., leaves -> leaf)
		if word[len(word)-3:] == "ves" {
			return word[:len(word)-3] + "f"
		}

		// -oes -> -o (e.g., heroes -> hero)
		if word[len(word)-3:] == "oes" {
			return word[:len(word)-2]
		}
	}

	if len(word) > 2 {
		// -es -> (e.g., boxes -> box, classes -> class)
		if word[len(word)-2:] == "es" {
			// Check if the base word ends with s, x, z, ch, sh
			base := word[:len(word)-2]
			if len(base) > 0 {
				lastChar := base[len(base)-1]
				if lastChar == 's' || lastChar == 'x' || lastChar == 'z' {
					return base
				}
				if len(base) > 1 {
					lastTwo := base[len(base)-2:]
					if lastTwo == "ch" || lastTwo == "sh" {
						return base
					}
				}
			}
			// Otherwise just remove 's'
			return word[:len(word)-1]
		}

		// -ss -> -ss (don't remove s from words like "class")
		if word[len(word)-2:] == "ss" {
			return word
		}

		// -s -> (e.g., users -> user)
		if word[len(word)-1:] == "s" {
			return word[:len(word)-1]
		}
	}

	return word
}
