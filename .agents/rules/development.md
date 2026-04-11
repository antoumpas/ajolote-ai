# Development Rules for ajolote-ai

## Every feature must work across all three commands

Every new feature added to ajolote must be exercised by all three entry points:

- **`ajolote init`** — the feature must work correctly on a fresh project (no `.agents/` yet), including importing existing tool state when available
- **`ajolote use <tool>`** — the feature must be reflected in the generated tool files after running `use`
- **`ajolote sync [<tool>]`** — the two-way sync must import the feature's data from tool config into `.agents/` (↑) and regenerate the tool files from the canonical config (↓)

If a feature only works in one direction or only one command, it is incomplete.

## Tests must cover all three commands

For every new feature, write integration or unit tests that exercise:

1. `init` — seeds the feature correctly (from existing tool config or from defaults)
2. `use` / `Generate()` — produces the correct tool output
3. `sync` / `Import()` — picks up new data from the tool config and merges it into `.agents/`

## Translator symmetry

When adding import (`Import()`) logic to one translator, consider whether the other translators should follow the same pattern. Prefer consistent behavior across all supported tools.
