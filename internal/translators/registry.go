package translators

import (
	"fmt"
	"strings"
)

// All returns all registered translators in a stable order.
func All() []Translator {
	return []Translator{
		&ClaudeTranslator{},
		&CursorTranslator{},
		&WindsurfTranslator{},
		&CopilotTranslator{},
		&ClineTranslator{},
		&AiderTranslator{},
	}
}

// Get returns the translator for the given tool name, or an error if unknown.
func Get(name string) (Translator, error) {
	for _, t := range All() {
		if t.Name() == name {
			return t, nil
		}
	}
	return nil, fmt.Errorf("unknown tool %q — run 'ajolote use <tool>' with one of: %s", name, Names())
}

// Names returns a comma-separated list of all supported tool names.
func Names() string {
	names := make([]string, 0, len(All()))
	for _, t := range All() {
		names = append(names, t.Name())
	}
	return strings.Join(names, ", ")
}
