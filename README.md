# Verkounter

A fast, concurrent word counting tool for tracking writing progress across multiple projects. Verkounter scans for Markdown files in designated folders, tracks daily word counts, and generates detailed statistics about your writing habits.

## Features

- 📁 **Project-based tracking**: Mark folders with `.verkount` files to track word counts
- 📚 **Series support**: Automatically groups projects by series and generates series-specific statistics
- 📊 **Smart change detection**: Only updates statistics when word counts actually change
- ⚡ **Concurrent processing**: Uses Go's goroutines for fast, parallel folder scanning
- 📈 **Detailed statistics**: View writing progress by day, week, month, year, and all-time
- 🔄 **Delta tracking**: Records actual words written each day, not just totals
- 📝 **YAML frontmatter aware**: Automatically strips YAML frontmatter from word counts

## Installation

### From Source

```bash
git clone https://github.com/yourusername/verkounter.git
cd verkounter
go build -o verkounter cmd/verkounter/main.go
```

### Optimized Build

For a smaller binary without debug symbols:

```bash
go build -ldflags="-s -w" -o verkounter cmd/verkounter/main.go
```

## Usage

### Basic Usage

Mark folders you want to track by creating a `.verkount` file:

```bash
touch ~/Documents/my-novel/.verkount
touch ~/Documents/blog-posts/.verkount
```

Run Verkounter to update word counts:

```bash
./verkounter
```

### View Statistics

Display detailed writing statistics:

```bash
./verkounter --stats
```

This shows:
- Today's writing progress
- Current week (Monday to Sunday) totals and daily average
- Past 30 days statistics
- Year-to-date progress
- Past 365 days overview
- Top 5 most productive writing days

## Project Structure

Verkounter expects your writing projects to be organized under `~/Documents`:

```
~/Documents/
├── Series-Name/           # Optional series folder
│   ├── Project-1/         # Individual project
│   │   ├── .verkount      # Marker file
│   │   └── *.md          # Markdown files to count
│   └── Project-2/
│       ├── .verkount
│       └── *.md
└── Standalone-Project/    # Projects can also be standalone
    ├── .verkount
    └── *.md
```

## Output Files

### Main Statistics File

`~/Documents/verkount_stats.yaml` - Contains all projects' daily word counts:

```yaml
2025-08-17:
  projects:
    Project-A: 1500
    Project-B: 2300
    My-Novel: 45000
  total: 48800
  delta: 1250  # Words written compared to previous entry
```

### Series Statistics Files

For each series folder, creates `verkount_series_stats.yaml`:

```yaml
2025-08-17:
  projects:
    Book-1: 45000
    Book-2: 38000
  total: 83000
  delta: 2000
```

## How It Works

1. **Scanning**: Recursively scans `~/Documents` for folders containing `.verkount` marker files
2. **Processing**: Reads all Markdown files in marked folders, stripping YAML frontmatter
3. **Counting**: Calculates word count using 6 characters = 1 word approximation
4. **Delta Calculation**: Compares with previous entry to determine words actually written
5. **Output**: Updates YAML files only when counts change, preserving writing history

## Architecture

- `cmd/verkounter/` - CLI entry point and command handling
- `internal/scanner/` - Directory scanning and .verkount detection
- `internal/processor/` - Markdown file processing and frontmatter stripping
- `internal/counter/` - Character/word counting logic
- `internal/output/` - YAML file generation and updates
- `internal/stats/` - Statistics calculation and display

## Development

### Running Tests

```bash
go test ./...
```

### Formatting Code

```bash
go fmt ./...
```

### Building

```bash
go build -o verkounter cmd/verkounter/main.go
```

## Configuration

Currently, Verkounter uses sensible defaults:
- Scans `~/Documents` directory
- Uses 6 characters per word ratio
- Creates output files in `~/Documents`
- Processes up to 4 folders concurrently

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) file for details

## Author

B. Wilson

## Acknowledgments

Built with Go and love for writers who want to track their progress without the bloat.