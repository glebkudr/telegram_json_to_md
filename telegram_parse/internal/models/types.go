package models

import "time"

// FileInfo represents information about a file being processed
type FileInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"modTime"`
	Status       string    `json:"status"` // "pending", "processing", "completed", "error"
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// Progress represents the current processing progress
type Progress struct {
	TotalFiles     int           `json:"totalFiles"`
	ProcessedFiles int           `json:"processedFiles"`
	CurrentFile    string        `json:"currentFile"`
	Percentage     float32       `json:"percentage"`
	IsActive       bool          `json:"isActive"`
	StartTime      time.Time     `json:"startTime"`
	EstimatedTime  time.Duration `json:"estimatedTime"`
}

// ProcessResult represents the result of processing operation
type ProcessResult struct {
	Success       bool          `json:"success"`
	TotalFiles    int           `json:"totalFiles"`
	SuccessCount  int           `json:"successCount"`
	ErrorCount    int           `json:"errorCount"`
	ProcessedSize int64         `json:"processedSize"`
	Duration      time.Duration `json:"duration"`
	Errors        []FileError   `json:"errors,omitempty"`
}

// FileError represents an error that occurred during file processing
type FileError struct {
	FilePath string `json:"filePath"`
	Error    string `json:"error"`
}

// ProcessOptions represents options for processing operation
type ProcessOptions struct {
	SourceDir      string `json:"sourceDir"`
	MaxConcurrency int    `json:"maxConcurrency"`
	IncludeSubdirs bool   `json:"includeSubdirs"`
}
