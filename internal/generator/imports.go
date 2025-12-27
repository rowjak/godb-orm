package generator

import (
	"sort"
	"strings"
)

// ImportManager tracks and generates required imports for generated code
type ImportManager struct {
	imports map[string]bool
}

// NewImportManager creates a new ImportManager instance
func NewImportManager() *ImportManager {
	return &ImportManager{
		imports: make(map[string]bool),
	}
}

// Add adds an import path to the manager
func (im *ImportManager) Add(importPath string) {
	if importPath != "" {
		im.imports[importPath] = true
	}
}

// AddMultiple adds multiple import paths
func (im *ImportManager) AddMultiple(importPaths ...string) {
	for _, path := range importPaths {
		im.Add(path)
	}
}

// Has checks if an import path is already added
func (im *ImportManager) Has(importPath string) bool {
	return im.imports[importPath]
}

// GetAll returns all import paths as a sorted slice
func (im *ImportManager) GetAll() []string {
	paths := make([]string, 0, len(im.imports))
	for path := range im.imports {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// Clear removes all imports
func (im *ImportManager) Clear() {
	im.imports = make(map[string]bool)
}

// Count returns the number of imports
func (im *ImportManager) Count() int {
	return len(im.imports)
}

// GenerateImportBlock generates the Go import block as a string
func (im *ImportManager) GenerateImportBlock() string {
	if len(im.imports) == 0 {
		return ""
	}

	paths := im.GetAll()

	// Separate standard library imports from third-party
	var stdLib, thirdParty []string
	for _, path := range paths {
		if isStdLib(path) {
			stdLib = append(stdLib, path)
		} else {
			thirdParty = append(thirdParty, path)
		}
	}

	var builder strings.Builder
	builder.WriteString("import (\n")

	// Write standard library imports first
	for _, path := range stdLib {
		builder.WriteString("\t\"")
		builder.WriteString(path)
		builder.WriteString("\"\n")
	}

	// Add blank line between std lib and third party if both exist
	if len(stdLib) > 0 && len(thirdParty) > 0 {
		builder.WriteString("\n")
	}

	// Write third-party imports
	for _, path := range thirdParty {
		builder.WriteString("\t\"")
		builder.WriteString(path)
		builder.WriteString("\"\n")
	}

	builder.WriteString(")")
	return builder.String()
}

// isStdLib checks if an import path is from the Go standard library
func isStdLib(path string) bool {
	// Standard library packages don't contain dots in their path
	// This is a simple heuristic that works for most cases
	stdLibPrefixes := []string{
		"archive/",
		"bufio",
		"bytes",
		"compress/",
		"container/",
		"context",
		"crypto/",
		"database/",
		"debug/",
		"embed",
		"encoding/",
		"errors",
		"expvar",
		"flag",
		"fmt",
		"go/",
		"hash/",
		"html/",
		"image/",
		"index/",
		"io",
		"log",
		"math",
		"mime/",
		"net/",
		"os",
		"path",
		"reflect",
		"regexp",
		"runtime",
		"sort",
		"strconv",
		"strings",
		"sync",
		"syscall",
		"testing",
		"text/",
		"time",
		"unicode",
		"unsafe",
	}

	for _, prefix := range stdLibPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") || strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// WellKnownImports contains common import paths used in generated code
var WellKnownImports = struct {
	Time       string
	Datatypes  string
	UUID       string
	GormDriver string
}{
	Time:       "time",
	Datatypes:  "gorm.io/datatypes",
	UUID:       "github.com/google/uuid",
	GormDriver: "gorm.io/gorm",
}
