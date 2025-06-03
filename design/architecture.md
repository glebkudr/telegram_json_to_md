# Telegram JSON to Markdown Parser - Architecture

## Overview
Desktop application for parsing Telegram JSON export files into clean Markdown format with background processing and progress tracking.

## Technology Stack
- **Backend**: Go 1.24.2+ with Wails v2
- **Frontend**: Vue 3 + Vite
- **JSON Parsing**: Go native encoding/json package
- **File Operations**: Standard Go file operations
- **Concurrency**: Goroutines for background processing

## Application Architecture

### Core Components

1. **Main Application (Go)**
   - Wails application entry point
   - Context management for app lifecycle

2. **File Service (Go)**
   - Directory selection and file discovery
   - JSON file validation and filtering
   - Batch processing coordination

3. **Parser Service (Go)**
   - JSON to Markdown conversion
   - Telegram data structure parsing
   - Progress tracking and reporting

4. **Frontend (Vue)**
   - Directory selection UI
   - Progress visualization
   - Processing status and results display

### Data Flow
```
User Input (Select Directory) 
    ↓
JSON File Discovery & Validation
    ↓
Background Processing (Goroutines)
    ↓
JSON → Markdown Conversion
    ↓
File Output & Progress Updates
    ↓
UI Updates via Context Bridge
```

## Telegram JSON Structure
- **result.json**: Main export file containing messages
- **Messages**: Array of message objects with text, media, etc.
- **Chat info**: Channel/group metadata
- **Media references**: Links to media files

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
│   ├── parser/              # JSON to Markdown parser
│   ├── fileops/             # File operations
│   ├── models/              # Data structures
│   └── telegram/            # Telegram-specific data structures
├── wails.json               # Wails configuration
└── design/                  # Documentation
```

## Key Features

### 1. Directory Processing
- Select source directory with JSON files
- Recursive file discovery
- File size and format validation

### 2. JSON Parsing
- Native Go JSON parsing
- Telegram data structure handling
- Message formatting and organization
- Media reference processing

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
- Stream processing for large JSON files (up to 50MB)
- Concurrent file processing
- Memory-efficient JSON parsing
- Progress update batching to prevent UI flooding

## Error Handling
- File access errors
- JSON parsing errors
- Invalid Telegram format errors
- Disk space validation
- Process cancellation
- Recovery from partial failures 