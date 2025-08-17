package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func main() {
	// Parse command line flags
	statsFlag := flag.Bool("stats", false, "Display writing statistics")
	flag.Parse()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}

	documentsPath := filepath.Join(homeDir, "Documents")

	// If --stats flag is provided, show statistics and exit
	if *statsFlag {
		showStatistics(documentsPath)
		return
	}

	fmt.Println("Scanning for .verkount folders...")
	folders, err := scanner.ScanForVerkountFolders(documentsPath)
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
	err = output.WriteStats(results, documentsPath)
	if err != nil {
		log.Fatalf("Error writing stats: %v", err)
	}

	// Write series-specific stats
	err = output.WriteSeriesStats(seriesResults, documentsPath)
	if err != nil {
		log.Fatalf("Error writing series stats: %v", err)
	}

	total := 0
	for _, count := range results {
		total += count
	}

	fmt.Printf("\nStats saved to ~/Documents/verkount_stats.yaml\n")
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

func showStatistics(documentsPath string) {
	// Load the stats file
	statsData, err := stats.LoadStats(documentsPath)
	if err != nil {
		log.Fatalf("Error loading stats: %v", err)
	}

	// Calculate and display statistics
	stats.CalculateStats(statsData)
}
