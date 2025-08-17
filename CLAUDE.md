# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Verkounter is a Go-based word counting tool that:
- Recursively scans ~/Documents for folders containing `.verkount` marker files
- Processes all Markdown files in those folders (recursively)
- Strips YAML frontmatter from Markdown files
- Counts characters and converts to word count (6 chars = 1 word average)
- Outputs results to a YAML file in ~/Documents with date, per-folder counts, and totals
- Uses Go's concurrency for parallel folder processing

## Build & Run Commands

```bash
# Build the application
go build -o verkounter cmd/verkounter/main.go

# Run the application
./verkounter

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code (if golangci-lint is installed)
golangci-lint run
```

## Status

✅ Application is working and tested. Successfully scans ~/Documents for .verkount markers, processes Markdown files, and outputs word counts to verkount_stats.yaml.

## Architecture

- `cmd/verkounter/main.go` - Entry point and CLI
- `internal/scanner/` - Directory scanning and .verkount detection
- `internal/processor/` - Markdown file processing and frontmatter stripping
- `internal/counter/` - Character/word counting logic
- `internal/output/` - YAML file generation and updating

## Key Design Decisions

- Concurrent processing using goroutines and channels for scalability
- 6 characters per word ratio for English text estimation
- YAML frontmatter detection using `---` delimiters
- Output file located at ~/Documents/verkount_stats.yaml
- Project names sanitized: spaces → hyphens, multiple hyphens collapsed
- Output format: nested hash with date as key, containing projects hash and total

## Output Format

```yaml
2025-08-17:
  projects:
    Project-A: 1500
    Project-B: 2300
    My-Notes: 850
  total: 4650
```