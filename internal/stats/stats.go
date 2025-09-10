package stats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

type DayStats struct {
	Projects map[string]int `yaml:"projects"`
	Total    int            `yaml:"total"`
	Delta    int            `yaml:"delta,omitempty"`
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
	return verkounterDir, nil
}

// LoadStats loads the verkount_stats.yaml file from XDG data directory
func LoadStats(documentsPath string) (StatsFile, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, fmt.Errorf("could not get data directory: %v", err)
	}

	statsFilePath := filepath.Join(dataDir, "verkount_stats.yaml")

	data, err := os.ReadFile(statsFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not read stats file: %v", err)
	}

	var stats StatsFile
	if err := yaml.Unmarshal(data, &stats); err != nil {
		return nil, fmt.Errorf("could not parse stats file: %v", err)
	}

	return stats, nil
}

// CalculateStats calculates statistics for various time periods
func CalculateStats(stats StatsFile) {
	now := time.Now()
	today := now.Format("2006-01-02")

	// Calculate daily deltas (words actually written each day)
	dailyDeltas := calculateDailyDeltas(stats)

	fmt.Print("\n=== Writing Statistics ===\n\n")

	// Today's stats
	if todayDelta, exists := dailyDeltas[today]; exists {
		fmt.Printf("Today (%s):\n", today)
		fmt.Printf("  Words written: %d\n\n", todayDelta)
	} else {
		fmt.Printf("Today (%s):\n", today)
		fmt.Println("  No words written yet")
		fmt.Println()
	}

	// This week's stats (Monday to Sunday)
	weekStart, weekEnd := getCurrentWeekRange(now)
	weekStats := calculatePeriodStatsFromDeltas(dailyDeltas, weekStart, weekEnd)
	daysInWeek := int(weekEnd.Sub(weekStart).Hours()/24) + 1
	if daysInWeek > 7 {
		daysInWeek = 7
	}

	fmt.Printf("This Week (Mon %s to Sun %s):\n", weekStart.Format("Jan 2"), weekEnd.Format("Jan 2"))
	fmt.Printf("  Total words: %d\n", weekStats.total)
	fmt.Printf("  Days with writing: %d/%d\n", weekStats.daysWithWriting, daysInWeek)
	if weekStats.daysWithWriting > 0 {
		fmt.Printf("  Daily average: %d words\n\n", weekStats.total/daysInWeek)
	} else {
		fmt.Print("\n")
	}

	// Past 30 days
	thirtyDaysAgo := now.AddDate(0, 0, -29) // -29 because we include today
	thirtyDayStats := calculatePeriodStatsFromDeltas(dailyDeltas, thirtyDaysAgo, now)

	fmt.Println("Past 30 Days:")
	fmt.Printf("  Total words: %d\n", thirtyDayStats.total)
	fmt.Printf("  Days with writing: %d/30\n", thirtyDayStats.daysWithWriting)
	if thirtyDayStats.daysWithWriting > 0 {
		fmt.Printf("  Daily average: %d words\n\n", thirtyDayStats.total/30)
	} else {
		fmt.Print("\n")
	}

	// Year to date
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	ytdStats := calculatePeriodStatsFromDeltas(dailyDeltas, yearStart, now)
	daysInYear := int(now.Sub(yearStart).Hours()/24) + 1

	fmt.Printf("Year to Date (%d):\n", now.Year())
	fmt.Printf("  Total words: %d\n", ytdStats.total)
	fmt.Printf("  Days with writing: %d/%d\n", ytdStats.daysWithWriting, daysInYear)
	if ytdStats.daysWithWriting > 0 {
		fmt.Printf("  Daily average: %d words\n\n", ytdStats.total/daysInYear)
	} else {
		fmt.Print("\n")
	}

	// Past 365 days
	yearAgo := now.AddDate(0, 0, -364) // -364 because we include today
	yearStats := calculatePeriodStatsFromDeltas(dailyDeltas, yearAgo, now)

	fmt.Println("Past 365 Days:")
	fmt.Printf("  Total words: %d\n", yearStats.total)
	fmt.Printf("  Days with writing: %d/365\n", yearStats.daysWithWriting)
	if yearStats.daysWithWriting > 0 {
		fmt.Printf("  Daily average: %d words\n\n", yearStats.total/365)
	} else {
		fmt.Print("\n")
	}

	// Most productive days
	showTopDaysFromDeltas(dailyDeltas, 5)
}

// getCurrentWeekRange returns Monday to Sunday of the current week
func getCurrentWeekRange(now time.Time) (time.Time, time.Time) {
	// Find the most recent Monday
	weekday := now.Weekday()
	daysFromMonday := int(weekday - time.Monday)
	if daysFromMonday < 0 {
		daysFromMonday += 7 // If today is Sunday
	}

	monday := now.AddDate(0, 0, -daysFromMonday)
	sunday := monday.AddDate(0, 0, 6)

	// Don't go beyond today
	if sunday.After(now) {
		sunday = now
	}

	return monday, sunday
}

type periodStats struct {
	total           int
	daysWithWriting int
}

// calculateDailyDeltas calculates the actual words written each day
// using the Delta field if available, or by comparing to previous day's total
func calculateDailyDeltas(stats StatsFile) map[string]int {
	deltas := make(map[string]int)
	
	// Get all dates and sort them
	var dates []string
	for date := range stats {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	
	// Use Delta field if available, otherwise calculate from totals
	for i, date := range dates {
		dayStats := stats[date]
		
		// If Delta field exists and is not the first entry, use it
		if dayStats.Delta > 0 || (dayStats.Delta == 0 && i > 0) {
			deltas[date] = dayStats.Delta
		} else if i == 0 {
			// First entry is baseline (0 words written that day)
			deltas[date] = 0
		} else {
			// Calculate delta from previous total (for old data without Delta field)
			prevTotal := stats[dates[i-1]].Total
			currTotal := dayStats.Total
			delta := currTotal - prevTotal
			if delta < 0 {
				delta = 0 // Handle cases where total might decrease
			}
			deltas[date] = delta
		}
	}
	
	return deltas
}

// calculatePeriodStatsFromDeltas calculates stats using daily deltas
func calculatePeriodStatsFromDeltas(deltas map[string]int, start, end time.Time) periodStats {
	result := periodStats{}
	
	for dateStr, delta := range deltas {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		
		// Check if date is within the period
		if (date.Equal(start) || date.After(start)) && (date.Equal(end) || date.Before(end)) {
			result.total += delta
			if delta > 0 {
				result.daysWithWriting++
			}
		}
	}
	
	return result
}

// showTopDaysFromDeltas shows the most productive writing days using daily deltas
func showTopDaysFromDeltas(deltas map[string]int, limit int) {
	type dayEntry struct {
		date  string
		words int
	}
	
	var days []dayEntry
	for date, wordsWritten := range deltas {
		if wordsWritten > 0 {
			days = append(days, dayEntry{date: date, words: wordsWritten})
		}
	}
	
	// Sort by words written descending
	sort.Slice(days, func(i, j int) bool {
		return days[i].words > days[j].words
	})
	
	fmt.Println("Most Productive Days:")
	count := limit
	if len(days) < limit {
		count = len(days)
	}
	
	for i := 0; i < count; i++ {
		date, _ := time.Parse("2006-01-02", days[i].date)
		fmt.Printf("  %d. %s: %d words\n", i+1, date.Format("Jan 2, 2006"), days[i].words)
	}
	
	if count == 0 {
		fmt.Println("  No writing days recorded yet")
	}
}
