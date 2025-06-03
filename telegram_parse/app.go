package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"telegram_parse/internal/fileops"
	"telegram_parse/internal/models"
	"telegram_parse/internal/parser"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	scanner *fileops.Scanner
	parser  *parser.JSONToMarkdown

	// Processing state
	mu              sync.Mutex
	isProcessing    bool
	currentProgress models.Progress
	cancelFunc      context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		scanner: fileops.NewScanner(),
		parser:  parser.NewJSONToMarkdown(),
	}
}

// OnStartup is called when the app starts up
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
}

// SelectDirectory opens a directory selection dialog
func (a *App) SelectDirectory() (string, error) {
	directory, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select directory with JSON files",
	})

	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}

	if directory == "" {
		return "", nil // User cancelled
	}

	// Validate directory
	err = a.scanner.ValidateDirectory(directory)
	if err != nil {
		return "", fmt.Errorf("invalid directory: %w", err)
	}

	return directory, nil
}

// ScanDirectory scans the selected directory for JSON files
func (a *App) ScanDirectory(dirPath string, includeSubdirs bool) ([]models.FileInfo, error) {
	files, err := a.scanner.ScanDirectory(dirPath, includeSubdirs)
	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, nil
}

// ProcessFiles starts processing JSON files to Markdown
func (a *App) ProcessFiles(options models.ProcessOptions) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isProcessing {
		return fmt.Errorf("processing is already in progress")
	}

	// Scan for files
	files, err := a.scanner.ScanDirectory(options.SourceDir, options.IncludeSubdirs)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no JSON files found in directory")
	}

	// Initialize progress
	a.currentProgress = models.Progress{
		TotalFiles:     len(files),
		ProcessedFiles: 0,
		CurrentFile:    "",
		Percentage:     0.0,
		IsActive:       true,
		StartTime:      time.Now(),
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancelFunc = cancel
	a.isProcessing = true

	// Start processing in background
	go a.processFilesBackground(ctx, files, options)

	return nil
}

// processFilesBackground handles file processing in background
func (a *App) processFilesBackground(ctx context.Context, files []models.FileInfo, options models.ProcessOptions) {
	defer func() {
		a.mu.Lock()
		a.isProcessing = false
		a.currentProgress.IsActive = false
		a.mu.Unlock()
	}()

	var (
		successCount  int
		errorCount    int
		errors        []models.FileError
		processedSize int64
	)

	// Process files with concurrency limit
	maxConcurrency := options.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // Default concurrency
	}

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, file := range files {
		select {
		case <-ctx.Done():
			// Processing was cancelled
			return
		default:
		}

		wg.Add(1)
		go func(fileInfo models.FileInfo, index int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Update current file progress
			a.updateProgress(fileInfo.Name, index)

			// Process the file
			outputPath := a.scanner.CreateOutputPath(fileInfo.Path)
			err := a.parser.ConvertFile(fileInfo.Path, outputPath)

			// Update results
			mu.Lock()
			if err != nil {
				errorCount++
				errors = append(errors, models.FileError{
					FilePath: fileInfo.Path,
					Error:    err.Error(),
				})
			} else {
				successCount++
				processedSize += fileInfo.Size
			}
			mu.Unlock()

		}(file, i)
	}

	wg.Wait()

	// Send final result
	result := models.ProcessResult{
		Success:       errorCount == 0,
		TotalFiles:    len(files),
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		ProcessedSize: processedSize,
		Duration:      time.Since(a.currentProgress.StartTime),
		Errors:        errors,
	}

	// Emit completion event
	runtime.EventsEmit(a.ctx, "processing-complete", result)
}

// updateProgress updates processing progress and emits event
func (a *App) updateProgress(currentFile string, processedIndex int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.currentProgress.ProcessedFiles = processedIndex + 1
	a.currentProgress.CurrentFile = currentFile
	a.currentProgress.Percentage = float32(a.currentProgress.ProcessedFiles) / float32(a.currentProgress.TotalFiles) * 100

	// Estimate remaining time
	elapsed := time.Since(a.currentProgress.StartTime)
	if a.currentProgress.ProcessedFiles > 0 {
		avgTimePerFile := elapsed / time.Duration(a.currentProgress.ProcessedFiles)
		remainingFiles := a.currentProgress.TotalFiles - a.currentProgress.ProcessedFiles
		a.currentProgress.EstimatedTime = avgTimePerFile * time.Duration(remainingFiles)
	}

	// Emit progress event
	runtime.EventsEmit(a.ctx, "processing-progress", a.currentProgress)
}

// GetProgress returns current processing progress
func (a *App) GetProgress() models.Progress {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentProgress
}

// CancelProcessing cancels ongoing processing
func (a *App) CancelProcessing() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isProcessing {
		return fmt.Errorf("no processing is in progress")
	}

	if a.cancelFunc != nil {
		a.cancelFunc()
	}

	return nil
}

// IsProcessing returns whether processing is currently active
func (a *App) IsProcessing() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.isProcessing
}
