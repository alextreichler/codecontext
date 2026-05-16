# CodeContext Project Instructions

This project follows specific conventions to maintain its goal of high-performance, token-efficient repository packing for AI agents.

## Core Mandates
- **Token Efficiency:** Every feature must prioritize "value per token." Avoid unnecessary boilerplate or verbose output formats.
- **XML Structure:** Always use `<f p="...">` tags for file wrapping. This is a reliable standard for LLM parsing.
- **Concurrency:** Performance-critical commands (bundle, search) must utilize the Go worker pool pattern to handle large repositories efficiently.

## Development Workflow
- **Surgical Changes:** Only update logic directly related to the task. Avoid unrelated refactoring.
- **Validation:** Always verify output with `codecontext` after making changes to ensures XML integrity and token accuracy.
- **New Symbols:** When adding support for new languages, ensure the regexes capture both the declaration and the body (for structs/interfaces) to maintain "Contextual Skeleton" quality.

## Agent Usage
- Start with `tree` or `index`.
- Use `skeleton` to understand API shapes.
- Use `bundle` or `extract` for implementation details.
