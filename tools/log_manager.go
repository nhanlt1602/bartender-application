package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LogFile represents a log file with metadata
type LogFile struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
}

// LogManager handles log file operations
type LogManager struct {
	LogDir     string
	MaxAge     int // days
	MaxHours   int // hours
	MaxMinutes int // minutes
	MaxBackups int
	MaxSize    int64 // bytes
}

// NewLogManager creates a new log manager
func NewLogManager(logDir string, maxAge, maxHours, maxMinutes, maxBackups int, maxSizeMB int) *LogManager {
	return &LogManager{
		LogDir:     logDir,
		MaxAge:     maxAge,
		MaxHours:   maxHours,
		MaxMinutes: maxMinutes,
		MaxBackups: maxBackups,
		MaxSize:    int64(maxSizeMB) * 1024 * 1024, // Convert MB to bytes
	}
}

// ListLogFiles lists all log files in the directory
func (lm *LogManager) ListLogFiles() ([]LogFile, error) {
	var files []LogFile

	entries, err := ioutil.ReadDir(lm.LogDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".log") {
			files = append(files, LogFile{
				Name:    entry.Name(),
				Path:    filepath.Join(lm.LogDir, entry.Name()),
				Size:    entry.Size(),
				ModTime: entry.ModTime(),
			})
		}
	}

	// Sort by modification time (oldest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.Before(files[j].ModTime)
	})

	return files, nil
}

// CleanupOldFiles removes log files older than MaxAge days
func (lm *LogManager) CleanupOldFiles() error {
	files, err := lm.ListLogFiles()
	if err != nil {
		return err
	}

	cutoffTime := time.Now().AddDate(0, 0, -lm.MaxAge).Add(-time.Duration(lm.MaxHours) * time.Hour).Add(-time.Duration(lm.MaxMinutes) * time.Minute)
	var deletedCount int

	for _, file := range files {
		if file.ModTime.Before(cutoffTime) {
			if err := os.Remove(file.Path); err != nil {
				fmt.Printf("Failed to delete old log file %s: %v\n", file.Name, err)
			} else {
				fmt.Printf("Deleted old log file: %s (age: %v)\n", file.Name, time.Since(file.ModTime))
				deletedCount++
			}
		}
	}

	fmt.Printf("Cleanup completed: deleted %d old log files\n", deletedCount)
	return nil
}

// CleanupExcessBackups removes excess backup files beyond MaxBackups
func (lm *LogManager) CleanupExcessBackups() error {
	files, err := lm.ListLogFiles()
	if err != nil {
		return err
	}

	if len(files) <= lm.MaxBackups {
		return nil
	}

	// Keep only the most recent MaxBackups files
	filesToDelete := files[:len(files)-lm.MaxBackups]
	var deletedCount int

	for _, file := range filesToDelete {
		if err := os.Remove(file.Path); err != nil {
			fmt.Printf("Failed to delete excess backup file %s: %v\n", file.Name, err)
		} else {
			fmt.Printf("Deleted excess backup file: %s\n", file.Name)
			deletedCount++
		}
	}

	fmt.Printf("Backup cleanup completed: deleted %d excess backup files\n", deletedCount)
	return nil
}

// GetLogStats returns statistics about log files
func (lm *LogManager) GetLogStats() error {
	files, err := lm.ListLogFiles()
	if err != nil {
		return err
	}

	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	fmt.Printf("Log Directory: %s\n", lm.LogDir)
	fmt.Printf("Total log files: %d\n", len(files))
	fmt.Printf("Total size: %.2f MB\n", float64(totalSize)/(1024*1024))
	fmt.Printf("Max age: %d days\n", lm.MaxAge)
	fmt.Printf("Max hours: %d\n", lm.MaxHours)
	fmt.Printf("Max minutes: %d\n", lm.MaxMinutes)
	fmt.Printf("Max backups: %d\n", lm.MaxBackups)
	fmt.Printf("Max size per file: %.2f MB\n", float64(lm.MaxSize)/(1024*1024))

	if len(files) > 0 {
		fmt.Printf("\nOldest file: %s (age: %v)\n", files[0].Name, time.Since(files[0].ModTime))
		fmt.Printf("Newest file: %s (age: %v)\n", files[len(files)-1].Name, time.Since(files[len(files)-1].ModTime))
	}

	return nil
}

func main() {
	// Example usage
	logDir := "./logs"
	//maxAge := 7     // 7 days
	//maxBackups := 5 // 5 backup files
	//maxSizeMB := 10 // 10 MB per file
	maxAge := 0     // 0 days
	maxHours := 0   // 0 hours
	maxMinutes := 2 // 0 minutes
	maxBackups := 5 // 5 backup files
	maxSizeMB := 10 // 10 MB per file

	lm := NewLogManager(logDir, maxAge, maxHours, maxMinutes, maxBackups, maxSizeMB)

	fmt.Println("=== Log Manager ===")

	// Show current stats
	if err := lm.GetLogStats(); err != nil {
		fmt.Printf("Error getting log stats: %v\n", err)
		return
	}

	fmt.Println("\n=== Cleanup Operations ===")

	// Cleanup old files
	if err := lm.CleanupOldFiles(); err != nil {
		fmt.Printf("Error cleaning up old files: %v\n", err)
	}

	// Cleanup excess backups
	if err := lm.CleanupExcessBackups(); err != nil {
		fmt.Printf("Error cleaning up excess backups: %v\n", err)
	}

	fmt.Println("\n=== Final Stats ===")
	if err := lm.GetLogStats(); err != nil {
		fmt.Printf("Error getting final stats: %v\n", err)
	}
}
