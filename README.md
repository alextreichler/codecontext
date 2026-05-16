# CodeContext

Machine-optimized repository packing for LLM context windows.

## Overview
`codecontext` is a high-performance CLI utility that aggregates local source code into a structured XML format. It is designed for machine consumption, prioritizing token density and structural clarity over human readability.

### Core Features
- **Structural Pruning:** `skeleton` mode extracts data structures and signatures while omitting implementation logic.
- **Surgical Extraction:** Target specific line ranges to minimize context bloat.
- **High-Throughput Aggregation:** Concurrent file processing via Go worker pools.
- **Deterministic XML:** Wraps files in `<f p="...">` tags for reliable model parsing.

## Commands

| Command | Description |
| :--- | :--- |
| `index` | Generates a symbol map with line numbers and signatures. |
| `tree` | Renders a directory structure visualization. |
| `skeleton` | Extracts type/struct bodies and doc comments. |
| `bundle` | Aggregates full file contents into XML. |
| `extract` | Retrieves a specific line range (e.g., `main.go 10:20`). |
| `search` | Concurrent grep with XML aggregation. |

## Configuration Options
- `--lines`: Prefixes output with line numbers.
- `--max-tokens <n>`: Truncates output at N tokens (default: 1M).
- `--verbose`: Includes character counts and processing metadata.

## Benchmarks
Results based on a 50-file, 5,000-line sample repository:

| Method | Size | Est. Tokens | Data Reduction |
| :--- | :--- | :--- | :--- |
| **Full Bundle** | 45.0 KB | ~11,250 | 0% |
| **Skeleton** | 7.5 KB | ~1,875 | 84% |
| **Index** | 2.8 KB | ~700 | 94% |

## Comparison: CodeContext vs Repomix
Benchmarks against Repomix (XML mode) on identical source material:

| Feature | Repomix (XML) | CodeContext |
| :--- | :--- | :--- |
| **Metadata Overhead** | ~15-20 KB (Headers/Tree) | < 1 KB |
| **Logic Compression** | ~23% (Tree-sitter) | ~85% (Skeleton) |
| **Execution Model** | Single-threaded | Multi-threaded Worker Pool |

## Ecosystem Integration

### Gemini CLI
Add to `GEMINI.md`:
```markdown
- Utilize `codecontext` for repository research and context gathering.
```

### Cline / Roo Code / Claude Dev
Add to `.claudecustominstructions`:
```text
Use `codecontext index .` for initial mapping.
Use `codecontext skeleton <path>` for structural analysis.
```

## Installation
```bash
go install github.com/alextreichler/codecontext@latest
```

## License
MIT
