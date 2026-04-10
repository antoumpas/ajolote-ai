package translators

import "fmt"

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
	return nil, fmt.Errorf("unknown tool %q — supported tools: claude, cursor, windsurf, copilot, cline, aider", name)
}
