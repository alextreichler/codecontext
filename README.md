# CodeContext 🚀

`codecontext` is a high-performance CLI tool designed to pack local repositories into an **agent-friendly XML format**. It maximizes "value per token" by providing surgical tools to explore codebases without overwhelming LLM context windows.

## Why use this?

Standard concatenation tools (like `cat`) or raw `grep` often provide too much or too little information. `codecontext` provides a middle ground:

- **Contextual Skeleton:** See struct fields and interface methods without implementation noise.
- **Surgical Extraction:** Grab specific line ranges to minimize context bloat.
- **Concurrent Processing:** Built in Go with a worker pool for ultra-fast bundling.
- **LLM-Optimized XML:** Uses `<f p="...">` tags which are highly reliable for model parsing.

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
