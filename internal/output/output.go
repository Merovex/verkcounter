package output

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bwilson/verkounter/internal/counter"
	"gopkg.in/yaml.v3"
)

type DayStats struct {
	Projects map[string]int `yaml:"projects"`
	Total    int            `yaml:"total"`
	Delta    int            `yaml:"delta,omitempty"` // Words written compared to previous entry
}

type StatsFile map[string]DayStats

// getDataDir returns the XDG data directory for verkounter
func getDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Check for XDG_DATA_HOME environment variable
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		dataHome = filepath.Join(homeDir, ".local", "share")
	}

	verkounterDir := filepath.Join(dataHome, "verkounter")

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(verkounterDir, 0755); err != nil {
		return "", err
	}

	return verkounterDir, nil
}

// migrateOldStats migrates stats files from ~/Documents to XDG data directory
func migrateOldStats() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dataDir, err := getDataDir()
	if err != nil {
		return err
	}

	// Check for old main stats file
	oldStatsPath := filepath.Join(homeDir, "Documents", "verkount_stats.yaml")
	newStatsPath := filepath.Join(dataDir, "verkount_stats.yaml")

	if _, err := os.Stat(oldStatsPath); err == nil {
		// Old file exists
		if _, err := os.Stat(newStatsPath); os.IsNotExist(err) {
			// New file doesn't exist, migrate
			fmt.Printf("Migrating stats from %s to %s\n", oldStatsPath, newStatsPath)
			if err := os.Rename(oldStatsPath, newStatsPath); err != nil {
				// If rename fails (e.g., cross-device), copy and delete
				data, readErr := os.ReadFile(oldStatsPath)
				if readErr != nil {
					return readErr
				}
				if writeErr := os.WriteFile(newStatsPath, data, 0644); writeErr != nil {
					return writeErr
				}
				if removeErr := os.Remove(oldStatsPath); removeErr != nil {
					fmt.Printf("Warning: Could not remove old stats file: %v\n", removeErr)
				}
			}
		}
	}

	return nil
}

func WriteStats(results map[string]int, outputPath string) error {
	// First, try to migrate old stats if they exist
	if err := migrateOldStats(); err != nil {
		fmt.Printf("Warning: Could not migrate old stats: %v\n", err)
	}

	dataDir, err := getDataDir()
	if err != nil {
		return err
	}

	statsFilePath := filepath.Join(dataDir, "verkount_stats.yaml")

	existingStats := make(StatsFile)

	data, err := os.ReadFile(statsFilePath)
	if err == nil {
		if err := yaml.Unmarshal(data, &existingStats); err != nil {
			fmt.Printf("Warning: Could not parse existing stats file: %v\n", err)
			existingStats = make(StatsFile)
		}
	}

	dateKey := time.Now().Format("2006-01-02")

	total := 0
	for _, count := range results {
		total += count
	}

	// Check if the most recent stats are identical to current results
	recentStats, _, found := getMostRecentStats(existingStats)
	if found {
		if statsAreEqual(recentStats.Projects, results) && recentStats.Total == total {
			fmt.Println("No changes in word counts - skipping update of main stats file")
			return nil
		}
	}

	// Calculate delta from most recent entry
	delta := 0
	if found {
		delta = total - recentStats.Total
		// First entry gets delta of 0 (it's the baseline)
	}

	existingStats[dateKey] = DayStats{
		Projects: results,
		Total:    total,
		Delta:    delta,
	}

	updatedData, err := yaml.Marshal(existingStats)
	if err != nil {
		return err
	}

	return os.WriteFile(statsFilePath, updatedData, 0644)
}

// migrateSeriesStats migrates series stats from Documents folders to XDG data directory
func migrateSeriesStats(seriesName string, documentsPath string) error {
	dataDir, err := getDataDir()
	if err != nil {
		return err
	}

	seriesDir := filepath.Join(dataDir, "series")
	if err := os.MkdirAll(seriesDir, 0755); err != nil {
		return err
	}

	oldStatsPath := filepath.Join(documentsPath, seriesName, "verkount_series_stats.yaml")
	newStatsPath := filepath.Join(seriesDir, seriesName+"_stats.yaml")

	if _, err := os.Stat(oldStatsPath); err == nil {
		// Old file exists
		if _, err := os.Stat(newStatsPath); os.IsNotExist(err) {
			// New file doesn't exist, migrate
			fmt.Printf("Migrating series stats from %s to %s\n", oldStatsPath, newStatsPath)
			data, readErr := os.ReadFile(oldStatsPath)
			if readErr != nil {
				return readErr
			}
			if writeErr := os.WriteFile(newStatsPath, data, 0644); writeErr != nil {
				return writeErr
			}
			if removeErr := os.Remove(oldStatsPath); removeErr != nil {
				fmt.Printf("Warning: Could not remove old series stats file: %v\n", removeErr)
			}
		}
	}

	return nil
}

// WriteSeriesStats writes stats for each series to a YAML file in the XDG data directory
func WriteSeriesStats(seriesResults map[string]map[string]int, scanPath string) error {
	dateKey := time.Now().Format("2006-01-02")

	dataDir, err := getDataDir()
	if err != nil {
		return err
	}

	seriesDir := filepath.Join(dataDir, "series")
	if err := os.MkdirAll(seriesDir, 0755); err != nil {
		return err
	}

	for seriesName, projects := range seriesResults {
		if seriesName == "" {
			// Skip projects that are not in a series folder
			continue
		}

		// Try to migrate old series stats if they exist (only if scanning Documents)
		homeDir, _ := os.UserHomeDir()
		documentsPath := filepath.Join(homeDir, "Documents")
		if scanPath == documentsPath {
			if err := migrateSeriesStats(seriesName, documentsPath); err != nil {
				fmt.Printf("Warning: Could not migrate series stats for %s: %v\n", seriesName, err)
			}
		}

		statsFilePath := filepath.Join(seriesDir, seriesName+"_stats.yaml")

		// Read existing stats if file exists
		existingStats := make(StatsFile)
		data, err := os.ReadFile(statsFilePath)
		if err == nil {
			if err := yaml.Unmarshal(data, &existingStats); err != nil {
				fmt.Printf("Warning: Could not parse existing series stats file for %s: %v\n", seriesName, err)
				existingStats = make(StatsFile)
			}
		}

		// Calculate total for this series
		total := 0
		sanitizedProjects := make(map[string]int)
		for projectName, count := range projects {
			sanitizedName := counter.SanitizeFolderName(projectName)
			sanitizedProjects[sanitizedName] = count
			total += count
		}

		// Check if the most recent stats are identical to current results
		recentStats, _, found := getMostRecentStats(existingStats)
		if found {
			if statsAreEqual(recentStats.Projects, sanitizedProjects) && recentStats.Total == total {
				fmt.Printf("No changes in word counts for series %s - skipping update\n", seriesName)
				continue
			}
		}

		// Calculate delta from most recent entry
		delta := 0
		if found {
			delta = total - recentStats.Total
			// First entry gets delta of 0 (it's the baseline)
		}

		// Update stats for today
		existingStats[dateKey] = DayStats{
			Projects: sanitizedProjects,
			Total:    total,
			Delta:    delta,
		}

		// Write updated stats
		updatedData, err := yaml.Marshal(existingStats)
		if err != nil {
			return fmt.Errorf("error marshaling series stats for %s: %v", seriesName, err)
		}

		if err := os.WriteFile(statsFilePath, updatedData, 0644); err != nil {
			return fmt.Errorf("error writing series stats for %s: %v", seriesName, err)
		}

		fmt.Printf("Series stats saved to %s\n", statsFilePath)
	}

	return nil
}

// statsAreEqual compares two maps of project stats to check if they're identical
func statsAreEqual(stats1, stats2 map[string]int) bool {
	if len(stats1) != len(stats2) {
		return false
	}

	for key, val1 := range stats1 {
		if val2, exists := stats2[key]; !exists || val1 != val2 {
			return false
		}
	}

	return true
}

// getMostRecentStats returns the most recent stats entry from a StatsFile
func getMostRecentStats(stats StatsFile) (DayStats, string, bool) {
	var mostRecentDate string

	for date := range stats {
		if date > mostRecentDate {
			mostRecentDate = date
		}
	}

	if mostRecentDate == "" {
		return DayStats{}, "", false
	}

	return stats[mostRecentDate], mostRecentDate, true
}
