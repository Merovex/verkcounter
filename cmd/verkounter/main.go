package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bwilson/verkounter/internal/counter"
	"github.com/bwilson/verkounter/internal/output"
	"github.com/bwilson/verkounter/internal/processor"
	"github.com/bwilson/verkounter/internal/scanner"
	"github.com/bwilson/verkounter/internal/stats"
)

type WorkResult struct {
	FolderName string
	SeriesName string
	WordCount  int
	Error      error
}

func printUsage() {
	fmt.Println(`Verkounter - A fast word counting tool that tracks writing progress across multiple projects
by scanning directories for .verkount marker files and processing Markdown content.

Usage:
  verkounter [directory]     Scan directory for .verkount projects (default: ~/Documents)
  verkounter --stats         Display writing statistics
  verkounter --help          Show this help message

Arguments:
  directory                  Directory to scan (optional, defaults to ~/Documents)
                            Examples: . (current dir), ~/Writing, /path/to/projects

Options:
  --stats                    Display detailed writing statistics
  --help, -h                 Show help information

Output files:
  Stats are stored in ~/.local/share/verkounter/
  - verkount_stats.yaml      Main statistics file with daily word counts
  - series/*_stats.yaml      Per-series statistics files

For more information, see: https://github.com/bwilson/verkounter`)
}

func main() {
	// Parse command line flags
	statsFlag := flag.Bool("stats", false, "Display writing statistics")
	helpFlag := flag.Bool("help", false, "Show help information")
	flag.BoolVar(helpFlag, "h", false, "Show help information (shorthand)")
	flag.Usage = printUsage
	flag.Parse()

	// Show help if requested
	if *helpFlag {
		printUsage()
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	// If --stats flag is provided, show statistics and exit
	if *statsFlag {
		showStatistics("")
		return
	}

	// Determine scan path from arguments or use default
	scanPath := filepath.Join(homeDir, "Documents") // default
	args := flag.Args()
	if len(args) > 0 {
		scanPath = args[0]
		
		// Handle current directory
		if scanPath == "." {
			scanPath, err = os.Getwd()
			if err != nil {
				log.Fatalf("Error getting current directory: %v", err)
			}
		} else if scanPath == ".." {
			// Handle parent directory
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Error getting current directory: %v", err)
			}
			scanPath = filepath.Dir(cwd)
		} else if strings.HasPrefix(scanPath, "~/") {
			// Handle tilde expansion
			scanPath = filepath.Join(homeDir, scanPath[2:])
		} else if !filepath.IsAbs(scanPath) {
			// Handle relative paths
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Error getting current directory: %v", err)
			}
			scanPath = filepath.Join(cwd, scanPath)
		}
	}

	// Verify the scan path exists
	if _, err := os.Stat(scanPath); os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", scanPath)
	}

	fmt.Printf("Scanning %s for .verkount folders...\n", scanPath)
	folders, err := scanner.ScanForVerkountFolders(scanPath)
	if err != nil {
		log.Fatalf("Error scanning folders: %v", err)
	}

	if len(folders) == 0 {
		fmt.Println("No folders with .verkount files found.")
		return
	}

	fmt.Printf("Found %d folders to process\n", len(folders))

	numWorkers := 4
	if len(folders) < numWorkers {
		numWorkers = len(folders)
	}

	folderChan := make(chan scanner.VerkountFolder, len(folders))
	resultChan := make(chan WorkResult, len(folders))

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(folderChan, resultChan, &wg)
	}

	for _, folder := range folders {
		folderChan <- folder
	}
	close(folderChan)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	results := make(map[string]int)
	seriesResults := make(map[string]map[string]int)
	errorCount := 0

	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("Error processing %s: %v\n", result.FolderName, result.Error)
			errorCount++
		} else {
			sanitizedName := counter.SanitizeFolderName(result.FolderName)
			results[sanitizedName] = result.WordCount
			fmt.Printf("  %s: %d words\n", sanitizedName, result.WordCount)

			// Track results by series
			if result.SeriesName != "" {
				if seriesResults[result.SeriesName] == nil {
					seriesResults[result.SeriesName] = make(map[string]int)
				}
				seriesResults[result.SeriesName][result.FolderName] = result.WordCount
			}
		}
	}

	if len(results) == 0 {
		fmt.Println("No results to save.")
		return
	}

	// Write overall stats
	err = output.WriteStats(results, scanPath)
	if err != nil {
		log.Fatalf("Error writing stats: %v", err)
	}

	// Write series-specific stats
	err = output.WriteSeriesStats(seriesResults, scanPath)
	if err != nil {
		log.Fatalf("Error writing series stats: %v", err)
	}

	total := 0
	for _, count := range results {
		total += count
	}

	// Determine XDG data directory for display message
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = "~/.local/share"
	} else {
		// Make it relative to home for display
		if homeDir, err := os.UserHomeDir(); err == nil {
			if rel, err := filepath.Rel(homeDir, dataHome); err == nil {
				dataHome = "~/" + rel
			}
		}
	}
	
	fmt.Printf("\nStats saved to %s/verkounter/verkount_stats.yaml\n", dataHome)
	fmt.Printf("Total words: %d\n", total)
	if errorCount > 0 {
		fmt.Printf("Errors encountered: %d\n", errorCount)
	}
}

func worker(folders <-chan scanner.VerkountFolder, results chan<- WorkResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for folder := range folders {
		content, err := processor.ProcessMarkdownFiles(folder.Path)
		if err != nil {
			results <- WorkResult{
				FolderName: folder.Name,
				SeriesName: folder.Series,
				Error:      err,
			}
			continue
		}

		wordCount := counter.CountWords(content)

		results <- WorkResult{
			FolderName: folder.Name,
			SeriesName: folder.Series,
			WordCount:  wordCount,
			Error:      nil,
		}
	}
}

func showStatistics(path string) {
	// Load the stats file from XDG data directory
	statsData, err := stats.LoadStats(path)
	if err != nil {
		log.Fatalf("Error loading stats: %v", err)
	}

	// Calculate and display statistics
	stats.CalculateStats(statsData)
}
