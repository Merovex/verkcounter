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

func WriteStats(results map[string]int, outputPath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	statsFilePath := filepath.Join(homeDir, "Documents", "verkount_stats.yaml")

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

// WriteSeriesStats writes stats for each series to a YAML file in the series folder
func WriteSeriesStats(seriesResults map[string]map[string]int, documentsPath string) error {
	dateKey := time.Now().Format("2006-01-02")

	for seriesName, projects := range seriesResults {
		if seriesName == "" {
			// Skip projects that are not in a series folder
			continue
		}

		seriesPath := filepath.Join(documentsPath, seriesName)
		statsFilePath := filepath.Join(seriesPath, "verkount_series_stats.yaml")

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
