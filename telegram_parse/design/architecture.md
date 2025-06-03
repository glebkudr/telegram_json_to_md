# Telegram HTML to Markdown Parser - Architecture

## Overview
Desktop application for parsing Telegram HTML export files into clean Markdown format with background processing and progress tracking.

## Technology Stack
- **Backend**: Go 1.24.2+ with Wails v2
- **Frontend**: Vue 3 + Vite
- **HTML Parsing**: Go html package + custom parser
- **File Operations**: Standard Go file operations
- **Concurrency**: Goroutines for background processing

## Application Architecture

### Core Components

1. **Main Application (Go)**
   - Wails application entry point
   - Context management for app lifecycle

2. **File Service (Go)**
   - Directory selection and file discovery
   - HTML file validation and filtering
   - Batch processing coordination

3. **Parser Service (Go)**
   - HTML to Markdown conversion
   - HTML tag cleaning and content extraction
   - Progress tracking and reporting

4. **Frontend (Vue)**
   - Directory selection UI
   - Progress visualization
   - Processing status and results display

### Data Flow
```
User Input (Select Directory) 
    ↓
File Discovery & Validation
    ↓
Background Processing (Goroutines)
    ↓
HTML → Markdown Conversion
    ↓
File Output & Progress Updates
    ↓
UI Updates via Context Bridge
```

## File Structure
```
telegram_parse/
├── app.go                    # Wails app struct and methods
├── main.go                   # Application entry point
├── build/                    # Wails build configuration
├── frontend/                 # Vue frontend
│   ├── dist/
│   ├── src/
│   │   ├── components/
│   │   ├── App.vue
│   │   └── main.js
│   ├── package.json
│   └── vite.config.js
├── internal/
│   ├── parser/              # HTML to Markdown parser
│   ├── fileops/             # File operations
│   └── models/              # Data structures
├── wails.json               # Wails configuration
└── design/                  # Documentation
```

## Key Features

### 1. Directory Processing
- Select source directory with HTML files
- Recursive file discovery
- File size and format validation

### 2. HTML Parsing
- Clean HTML tag removal
- Content extraction
- Telegram-specific formatting preservation
- Character encoding handling

### 3. Background Processing
- Goroutine-based parallel processing
- Progress tracking per file
- Error handling and reporting
- Graceful cancellation support

### 4. User Interface
- Directory selection dialog
- Real-time progress bar
- Processing status display
- Error notifications
- Results summary

## Performance Considerations
- Stream processing for large files (up to 50MB)
- Concurrent file processing
- Memory-efficient HTML parsing
- Progress update batching to prevent UI flooding

## Error Handling
- File access errors
- HTML parsing errors
- Disk space validation
- Process cancellation
- Recovery from partial failures 