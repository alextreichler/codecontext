# CodeContext - Claude Instructions

This file provides guidance for Claude-based agents (Claude Dev, Cline, etc.) to maintain consistency and efficiency in the `CodeContext` project.

## Project Goal
`codecontext` is a high-performance Go-based CLI tool designed to pack repositories into token-efficient XML for LLM context.

## Build & Test Commands
- **Build:** `go build -o codecontext main.go`
- **Lint:** `go fmt ./...`
- **Run (Local):** `./codecontext <command> [args]`

## Repository Structure
- `main.go`: Single-file core implementation (CLI, concurrent processing, XML formatting).
- `go.mod`: Project dependencies (primarily `go-gitignore`).

## Development Conventions
1. **Prefer `codecontext` for Research:** When exploring this (or any) codebase, use `./codecontext index .` or `./codecontext skeleton .` first. It is significantly more token-efficient than reading full files.
2. **Surgical Edits:** Only modify code strictly related to the requested feature or fix.
3. **XML Integrity:** Ensure all file output is wrapped in `<f p="path">` tags. Use `xmlEscape` for all dynamic path/content output.
4. **Concurrency:** When adding commands that scan multiple files, implement them using the established worker pool pattern in `runBundle` or `runSearch`.

## Preferred Agent Workflow
1. **Index:** Use `./codecontext index .` to find symbols.
2. **Skeleton:** Use `./codecontext skeleton <path>` to understand API shapes/structs.
3. **Extract:** Use `./codecontext extract <file> <range>` for targeted code reading.
4. **Bundle:** Use `./codecontext bundle <path>` only when full implementation context is required.
