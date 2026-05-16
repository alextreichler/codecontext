# CodeContext 🚀

**Note: This tool is designed for AI agents, not human readability.** It generates structured, high-density context that allows LLMs to understand codebases while minimizing token waste.

`codecontext` is a high-performance CLI tool designed to pack local repositories into an **agent-friendly XML format**. It maximizes "value per token" by providing surgical tools to explore codebases without overwhelming LLM context windows.

## Why use this?

Standard concatenation tools (like `cat`) or raw `grep` often provide too much or too little information. `codecontext` provides a middle ground:

- **Contextual Skeleton:** See struct fields and interface methods without implementation noise.
- **Surgical Extraction:** Grab specific line ranges to minimize context bloat.
- **Concurrent Processing:** Built in Go with a worker pool for ultra-fast bundling.
- **LLM-Optimized XML:** Uses `<f p="...">` tags which are highly reliable for model parsing.

## Ecosystem Integration

To give your AI agent "superpowers," add `codecontext` to its configuration:

### 1. Gemini CLI
Add this to your global `~/.gemini/GEMINI.md` or local `GEMINI.md`:
```markdown
- Always prefer using the `codecontext` CLI tool for codebase research, indexing, and gathering context.
```

### 2. Cline / Roo Code / Claude Dev
Add this to your `.claudecustominstructions`:
```text
Use `codecontext index .` to map the project before reading files.
Use `codecontext skeleton <path>` to see data structures.
```

### 3. Custom Agents / MCP
If you are building a custom agent, you can wrap `codecontext` as a tool:
- **Input:** Search query or file path.
- **Output:** The XML-wrapped response from `codecontext`.

## Use Case: Debugging & Troubleshooting

`codecontext` is exceptionally powerful when you need an AI to help diagnose an issue in a **running system**:

1.  **Map the Flow:** Start with `codecontext index .` to let the AI see the entire project's entry points and handlers.
2.  **Inspect State:** Use `codecontext skeleton` on the models or state-management files to help the AI understand how data is structured and where it might be corrupted.
3.  **Trace Errors:** If you have a stack trace, give it to the AI and have it use `codecontext extract` to look at the exact line ranges mentioned in the trace without the noise of the rest of the file.
4.  **Identify Side Effects:** Use `codecontext search` for specific environment variables or database keys to find every place the system interacts with external state.

## Installation

```bash
go install github.com/alextreichler/codecontext@latest
```

## Commands

| Command | Description | Best For... |
| :--- | :--- | :--- |
| **`index`** | Compact symbol map (lines numbers + signatures). | Initial discovery. |
| **`tree`** | Visual directory structure. | Understanding layout. |
| **`skeleton`** | Type/Struct bodies + doc comments. | Understanding APIs/Contracts. |
| **`bundle`** | Full file contents in XML. | Refactoring or debugging. |
| **`extract`** | Specific line range (e.g., `main.go 10:20`). | Surgical edits. |
| **`search`** | Concurrent grep + XML bundling. | Finding logic across modules. |

## Options

- `--lines`: Prefix each line with its line number (essential for editing).
- `--max-tokens <n>`: Safety limit to prevent context overflows (default 1M).
- `--verbose`: Include character/token estimates and mode metadata.

## Agent Best Practices

1.  **Discover:** Start with `tree` or `index`.
2.  **Understand:** Use `skeleton` on relevant directories to see the "shape" of the code.
3.  **Read:** Use `bundle` or `extract` only on the specific files you need to modify.
4.  **Verify:** Use `search` to ensure your changes don't break dependencies elsewhere.


## License

MIT
