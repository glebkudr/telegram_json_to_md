package fileops

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"telegram_parse/internal/models"
)

// Scanner handles file system operations
type Scanner struct{}

// NewScanner creates a new file scanner
func NewScanner() *Scanner {
	return &Scanner{}
}

// ScanDirectory scans directory for JSON files
func (s *Scanner) ScanDirectory(dirPath string, includeSubdirs bool) ([]models.FileInfo, error) {
	var files []models.FileInfo

	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip subdirectories if not requested
		if !includeSubdirs && path != dirPath && d.IsDir() {
			return filepath.SkipDir
		}

		// Process only JSON files
		if !d.IsDir() && s.isJSONFile(path) {
			info, err := d.Info()
			if err != nil {
				return err
			}

			fileInfo := models.FileInfo{
				Path:    path,
				Name:    d.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Status:  "pending",
			}

			files = append(files, fileInfo)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, nil
}

// isJSONFile checks if file has JSON extension
func (s *Scanner) isJSONFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".json"
}

// ValidateDirectory checks if directory exists and is accessible
func (s *Scanner) ValidateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("cannot access directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// GetFileSize returns file size in bytes
func (s *Scanner) GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, fmt.Errorf("cannot get file size: %w", err)
	}
	return info.Size(), nil
}

// CreateOutputPath creates output markdown file path
func (s *Scanner) CreateOutputPath(jsonPath string) string {
	dir := filepath.Dir(jsonPath)
	name := strings.TrimSuffix(filepath.Base(jsonPath), filepath.Ext(jsonPath))
	return filepath.Join(dir, name+".md")
}

// CheckDiskSpace checks if there's enough disk space for processing
func (s *Scanner) CheckDiskSpace(path string, requiredBytes int64) error {
	// Get disk usage for the path
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot check disk space: %w", err)
	}

	// For now, we'll just check if the file exists and is accessible
	// In a more sophisticated implementation, we could use platform-specific
	// system calls to get actual disk space
	_ = stat

	return nil
}
