# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Verkounter is a Go-based word counting tool that:
- Recursively scans directories for folders containing `.verkount` marker files
- Default scan directory is ~/Documents, but can scan any specified directory
- Processes all Markdown files in those folders (recursively)
- Strips YAML frontmatter from Markdown files
- Counts characters and converts to word count (6 chars = 1 word average)
- Stores stats in `~/.local/share/verkounter/` following XDG Base Directory Specification
- Uses Go's concurrency for parallel folder processing

## Build & Run Commands

```bash
# Build the application
go build -o verkounter cmd/verkounter/main.go

# Run the application
./verkounter              # Scan ~/Documents (default)
./verkounter .            # Scan current directory
./verkounter ~/Writing    # Scan specific directory

# View statistics
./verkounter --stats

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint code (if golangci-lint is installed)
golangci-lint run
```

## Status

✅ Application is working and tested. Successfully scans directories for .verkount markers, processes Markdown files, and stores word counts in XDG data directory (~/.local/share/verkounter/).

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
- Stats stored in XDG data directory: `~/.local/share/verkounter/`
- Supports scanning any directory via positional argument (defaults to ~/Documents)
- Project names sanitized: spaces → hyphens, multiple hyphens collapsed
- Output format: nested hash with date as key, containing projects hash, total, and delta

## Output Format

```yaml
2025-08-17:
  projects:
    Project-A: 1500
    Project-B: 2300
    My-Notes: 850
  total: 4650
```

## Lessons Learned

### General Software Engineering Principles

- **Follow platform conventions**: Use XDG Base Directory Specification on Linux rather than arbitrary locations (e.g., ~/.local/share/ for data, not ~/Documents)
- **Migration before breaking changes**: When changing data storage locations, implement automatic migration to preserve user data
- **Positional args over flags for primary targets**: Use `tool [target]` pattern like Unix tools (ls, grep) rather than `tool --dir=target`
- **Centralized data, distributed sources**: Store output/stats in one canonical location regardless of where input is scanned from
- **Smart updates**: Only write files when data actually changes to avoid unnecessary disk I/O and preserve timestamps
- **Progressive disclosure in help**: Show concise summary first (25-35 words), then usage patterns, then detailed options
- **Handle path expansion properly**: Support `.`, `..`, `~/`, relative paths, and absolute paths for maximum flexibility
- **Fail gracefully with helpful errors**: Check if paths exist before processing and provide clear error messages