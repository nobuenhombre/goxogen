package domainapp

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/nobuenhombre/suikat/pkg/ge"
)

// dedupFunctions removes duplicate Go function/method declarations from .xo.go,
// .xouid.go, and *-repo.xo.go files. When multiple database indexes with the same
// columns generate identical Go functions (same name and receiver), only the first
// occurrence is kept and the rest are removed.
func (d *AppDomain) dedupFunctions(dir string) error {
	// Process .xo.go and .xouid.go files (excluding -repo.xo.go)
	err := d.dedupInFiles(dir, []string{"*.xo.go", "*.xouid.go"}, "-repo.xo.go")
	if err != nil {
		return ge.Pin(err)
	}

	// Process *-repo.xo.go files (repo methods can also be duplicated)
	err = d.dedupInFiles(dir, []string{"*-repo.xo.go"}, "")
	if err != nil {
		return ge.Pin(err)
	}

	return nil
}

// dedupInFiles finds duplicate function declarations in matching files and removes them.
func (d *AppDomain) dedupInFiles(dir string, patterns []string, excludeSuffix string) error {
	var files []string

	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return ge.Pin(err)
		}
		files = append(files, matches...)
	}

	sort.Strings(files)

	seen := make(map[string]bool) // function key -> already seen (global across files)

	for _, file := range files {
		base := filepath.Base(file)
		if excludeSuffix != "" && strings.HasSuffix(base, excludeSuffix) {
			continue
		}

		data, err := os.ReadFile(file)
		if err != nil {
			return ge.Pin(fmt.Errorf("reading %s: %w", file, err))
		}

		content := string(data)
		if !strings.Contains(content, "func ") {
			continue
		}

		deduped := removeDuplicateFuncs(content, seen)

		if deduped != content {
			err = os.WriteFile(file, []byte(deduped), 0644)
			if err != nil {
				return ge.Pin(fmt.Errorf("writing %s: %w", file, err))
			}
			log.Printf("[xo] Removed duplicate functions from %s", base)
		}
	}

	return nil
}

// funcRange describes a function declaration's span in the source file.
type funcRange struct {
	start int // line index of "func "
	end   int // line index after the closing "}" (exclusive)
	key   string
}

// removeDuplicateFuncs parses Go source and removes duplicate function/method declarations.
//
// Strategy: two-pass approach so duplicates can be anywhere in the file (not just consecutive):
//  1. First pass: parse ALL function boundaries and collect their keys
//  2. Second pass: build result — for each function, keep only the first occurrence per key
//
// The `seen` map is GLOBAL across files: once a function key is seen in any file,
// subsequent occurrences in any other file are also removed. This is correct because
// Go compiles all files in a package together — duplicate names across files also fail.
func removeDuplicateFuncs(content string, seen map[string]bool) string {
	lines := strings.Split(content, "\n")

	// ── Pass 1: parse all function boundaries ──
	var funcs []funcRange

	i := 0
	for i < len(lines) {
		trimmed := strings.TrimLeftFunc(lines[i], unicode.IsSpace)
		if strings.HasPrefix(trimmed, "func ") {
			start := i

			// Find the line containing '{' (signature may span multiple lines)
			j := i + 1
			for j < len(lines) && !strings.Contains(lines[j-1], "{") {
				j++
			}
			// Now lines[j-1] contains '{' — the opening brace of the function body

			// Track brace depth from the '{' line to find the matching '}'
			depth := strings.Count(lines[j-1], "{") - strings.Count(lines[j-1], "}")
			k := j
			for k < len(lines) && depth > 0 {
				depth += strings.Count(lines[k], "{") - strings.Count(lines[k], "}")
				k++
			}
			// depth == 0 at lines[k-1]; function spans lines[start:k]

			// Extract key from signature lines (everything before the '{' line)
			sigLines := lines[start:j]
			key := extractFuncKey(sigLines)

			funcs = append(funcs, funcRange{start: start, end: k, key: key})
			i = k
			continue
		}
		i++
	}

	// ── Pass 2: determine which functions to keep ──
	keep := make([]bool, len(funcs))
	for idx, fn := range funcs {
		if fn.key == "" {
			keep[idx] = true // functions without a key are always kept
			continue
		}
		if seen[fn.key] {
			keep[idx] = false
		} else {
			seen[fn.key] = true
			keep[idx] = true
		}
	}

	// ── Pass 3: build result ──
	var result []string
	lastEnd := 0

	for idx, fn := range funcs {
		if keep[idx] {
			// Copy non-function lines between the previous function and this one
			result = append(result, lines[lastEnd:fn.start]...)
			// Copy the function itself (signature + body)
			result = append(result, lines[fn.start:fn.end]...)
			lastEnd = fn.end
		} else {
			// Duplicate: skip the gap and the function body.
			// BUT: preserve content before the first function
			// (package declaration, imports) even if the first function is a duplicate.
			if lastEnd == 0 {
				result = append(result, lines[0:fn.start]...)
			}
			lastEnd = fn.end
		}
	}

	// Copy any trailing lines after the last function
	result = append(result, lines[lastEnd:]...)

	return strings.Join(result, "\n")
}

// extractFuncKey extracts a unique key from Go function/method signature lines.
// For top-level functions: returns the function name.
// For methods: returns "*ReceiverType.MethodName".
func extractFuncKey(sigLines []string) string {
	if len(sigLines) == 0 {
		return ""
	}

	// Join all signature lines into one compact string
	var sb strings.Builder
	for _, line := range sigLines {
		sb.WriteString(strings.TrimSpace(line))
		sb.WriteString(" ")
	}
	sig := strings.TrimSpace(sb.String())

	if !strings.HasPrefix(sig, "func ") {
		return ""
	}

	// Remove "func " prefix
	rest := strings.TrimSpace(sig[5:])

	if strings.HasPrefix(rest, "(") {
		// Method: func (receiver *Type) MethodName(...)
		depth := 0
		receiverType := ""
		methodName := ""

		for i, c := range rest {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					recv := strings.TrimSpace(rest[1:i])
					receiverType = extractReceiverType(recv)

					after := strings.TrimSpace(rest[i+1:])
					methodName = extractIdentifier(after)
					break
				}
			}
		}

		if receiverType != "" && methodName != "" {
			return receiverType + "." + methodName
		}
		return ""
	}

	// Top-level function: extract name before '('
	return extractIdentifier(rest)
}

// extractReceiverType extracts the underlying type from a receiver declaration.
// e.g., "repo *TaskRepository" -> "*TaskRepository"
func extractReceiverType(recv string) string {
	recv = strings.TrimSpace(recv)
	fields := strings.Fields(recv)
	if len(fields) > 0 {
		return fields[len(fields)-1]
	}
	return ""
}

// extractIdentifier extracts the first identifier before '(' from a string.
// e.g., "GetTasksByPerformerCloseDate(db ..." -> "GetTasksByPerformerCloseDate"
func extractIdentifier(s string) string {
	s = strings.TrimSpace(s)
	for i, c := range s {
		if c == '(' || c == ' ' || c == '\t' || c == '\n' {
			if i > 0 {
				return s[:i]
			}
			return ""
		}
	}
	return s
}
