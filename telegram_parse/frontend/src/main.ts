import './style.css';
import './app.css';

import {
    SelectDirectory,
    ScanDirectory,
    ProcessFiles,
    CancelProcessing
} from '../wailsjs/go/main/App';

import { EventsOn } from '../wailsjs/runtime/runtime';

// Application state
interface AppState {
    selectedDirectory: string;
    files: any[];
    isProcessing: boolean;
    progress: any;
    results: any;
    showResults: boolean;
    includeSubdirs: boolean;
    maxConcurrency: number;
}

const state: AppState = {
    selectedDirectory: '',
    files: [],
    isProcessing: false,
    progress: {
        totalFiles: 0,
        processedFiles: 0,
        currentFile: '',
        percentage: 0,
        isActive: false,
        estimatedTime: 0
    },
    results: null,
    showResults: false,
    includeSubdirs: false,
    maxConcurrency: 4
};

// DOM elements
let selectDirBtn: HTMLButtonElement;
let selectedDirSpan: HTMLSpanElement;
let fileCountSpan: HTMLSpanElement;
let processBtn: HTMLButtonElement;
let cancelBtn: HTMLButtonElement;
let progressContainer: HTMLDivElement;
let progressBar: HTMLDivElement;
let progressText: HTMLSpanElement;
let currentFileSpan: HTMLSpanElement;
let resultsContainer: HTMLDivElement;
let includeSubdirsCheckbox: HTMLInputElement;
let maxConcurrencyInput: HTMLInputElement;

// Initialize the application
document.querySelector('#app')!.innerHTML = `
    <div class="container">
        <header class="header">
            <h1>📱 Telegram JSON to Markdown Parser</h1>
            <p>Convert Telegram JSON exports to clean Markdown files</p>
        </header>

        <div class="main-content">
            <!-- Directory Selection -->
            <div class="section">
                <h2>📁 Select Directory</h2>
                <div class="directory-selector">
                    <button id="selectDirBtn" class="btn btn-primary">
                        Choose Directory
                    </button>
                    <span id="selectedDir" class="directory-path">
                        No directory selected
                    </span>
                </div>
                <p class="help-text">
                    Select a directory containing Telegram JSON export files (result.json)
                </p>
            </div>

            <!-- Options -->
            <div class="section">
                <h2>⚙️ Options</h2>
                <div class="options">
                    <label class="checkbox-label">
                        <input type="checkbox" id="includeSubdirs">
                        <span class="checkmark"></span>
                        Include subdirectories
                    </label>
                    
                    <div class="input-group">
                        <label for="maxConcurrency">Max concurrent files:</label>
                        <input type="number" id="maxConcurrency" min="1" max="10" value="4" class="input-number">
                    </div>
                </div>
            </div>

            <!-- File Information -->
            <div class="section" id="fileSection" style="display: none;">
                <h2>📋 Found Files</h2>
                <div class="file-info">
                    <span id="fileCount" class="file-count">0 JSON files found</span>
                </div>
            </div>

            <!-- Processing Controls -->
            <div class="section" id="controlsSection" style="display: none;">
                <div class="controls">
                    <button id="processBtn" class="btn btn-success">
                        🚀 Start Processing
                    </button>
                    <button id="cancelBtn" class="btn btn-danger" style="display: none;">
                        ⏹️ Cancel
                    </button>
                </div>
            </div>

            <!-- Progress -->
            <div class="section" id="progressSection" style="display: none;">
                <h2>⏳ Processing Progress</h2>
                <div class="progress-container">
                    <div class="progress-bar">
                        <div id="progressBar" class="progress-fill"></div>
                    </div>
                    <div class="progress-info">
                        <span id="progressText" class="progress-text">0%</span>
                        <span id="currentFile" class="current-file"></span>
                    </div>
                </div>
            </div>

            <!-- Results -->
            <div class="section" id="resultsSection" style="display: none;">
                <h2>✅ Results</h2>
                <div id="resultsContainer" class="results">
                    <!-- Results will be inserted here -->
                </div>
            </div>
        </div>
    </div>
`;

// Get DOM elements
selectDirBtn = document.getElementById('selectDirBtn') as HTMLButtonElement;
selectedDirSpan = document.getElementById('selectedDir') as HTMLSpanElement;
fileCountSpan = document.getElementById('fileCount') as HTMLSpanElement;
processBtn = document.getElementById('processBtn') as HTMLButtonElement;
cancelBtn = document.getElementById('cancelBtn') as HTMLButtonElement;
progressContainer = document.getElementById('progressSection') as HTMLDivElement;
progressBar = document.getElementById('progressBar') as HTMLDivElement;
progressText = document.getElementById('progressText') as HTMLSpanElement;
currentFileSpan = document.getElementById('currentFile') as HTMLSpanElement;
resultsContainer = document.getElementById('resultsContainer') as HTMLDivElement;
includeSubdirsCheckbox = document.getElementById('includeSubdirs') as HTMLInputElement;
maxConcurrencyInput = document.getElementById('maxConcurrency') as HTMLInputElement;

// Event listeners
selectDirBtn.addEventListener('click', selectDirectory);
processBtn.addEventListener('click', startProcessing);
cancelBtn.addEventListener('click', cancelProcessing);

includeSubdirsCheckbox.addEventListener('change', (e) => {
    state.includeSubdirs = (e.target as HTMLInputElement).checked;
    if (state.selectedDirectory) {
        scanDirectory();
    }
});

maxConcurrencyInput.addEventListener('change', (e) => {
    state.maxConcurrency = parseInt((e.target as HTMLInputElement).value);
});

// Listen for backend events
EventsOn('processing-progress', (progress: any) => {
    updateProgress(progress);
});

EventsOn('processing-complete', (results: any) => {
    processingComplete(results);
});

// Functions
async function selectDirectory() {
    try {
        const directory = await SelectDirectory();
        if (directory) {
            state.selectedDirectory = directory;
            selectedDirSpan.textContent = directory;
            selectedDirSpan.className = 'directory-path selected';
            await scanDirectory();
        }
    } catch (error) {
        console.error('Error selecting directory:', error);
        alert('Error selecting directory: ' + error);
    }
}

async function scanDirectory() {
    if (!state.selectedDirectory) return;
    
    try {
        const files = await ScanDirectory(state.selectedDirectory, state.includeSubdirs);
        state.files = files;
        
        fileCountSpan.textContent = `${files.length} JSON files found`;
        
        const fileSection = document.getElementById('fileSection')!;
        const controlsSection = document.getElementById('controlsSection')!;
        
        if (files.length > 0) {
            fileSection.style.display = 'block';
            controlsSection.style.display = 'block';
            processBtn.disabled = false;
        } else {
            fileSection.style.display = 'block';
            controlsSection.style.display = 'none';
            alert('No JSON files found in the selected directory.');
        }
    } catch (error) {
        console.error('Error scanning directory:', error);
        alert('Error scanning directory: ' + error);
    }
}

async function startProcessing() {
    if (!state.selectedDirectory || state.files.length === 0) return;
    
    try {
        const options = {
            sourceDir: state.selectedDirectory,
            maxConcurrency: state.maxConcurrency,
            includeSubdirs: state.includeSubdirs
        };
        
        await ProcessFiles(options);
        
        state.isProcessing = true;
        processBtn.style.display = 'none';
        cancelBtn.style.display = 'inline-block';
        progressContainer.style.display = 'block';
        
        // Hide results section
        document.getElementById('resultsSection')!.style.display = 'none';
        
    } catch (error) {
        console.error('Error starting processing:', error);
        alert('Error starting processing: ' + error);
    }
}

async function cancelProcessing() {
    try {
        await CancelProcessing();
        resetProcessingState();
    } catch (error) {
        console.error('Error cancelling processing:', error);
    }
}

function updateProgress(progress: any) {
    state.progress = progress;
    
    const percentage = Math.round(progress.percentage);
    progressBar.style.width = `${percentage}%`;
    progressText.textContent = `${percentage}% (${progress.processedFiles}/${progress.totalFiles})`;
    
    if (progress.currentFile) {
        currentFileSpan.textContent = `Processing: ${progress.currentFile}`;
    }
    
    if (progress.estimatedTime > 0) {
        const remainingSeconds = Math.ceil(progress.estimatedTime / 1000000000); // Convert nanoseconds to seconds
        const remainingText = remainingSeconds > 60 
            ? `${Math.ceil(remainingSeconds / 60)} minutes remaining`
            : `${remainingSeconds} seconds remaining`;
        currentFileSpan.textContent += ` - ${remainingText}`;
    }
}

function processingComplete(results: any) {
    state.results = results;
    state.isProcessing = false;
    resetProcessingState();
    
    // Show results
    displayResults(results);
    document.getElementById('resultsSection')!.style.display = 'block';
}

function resetProcessingState() {
    processBtn.style.display = 'inline-block';
    cancelBtn.style.display = 'none';
    progressContainer.style.display = 'none';
    
    // Reset progress
    progressBar.style.width = '0%';
    progressText.textContent = '0%';
    currentFileSpan.textContent = '';
}

function displayResults(results: any) {
    const duration = Math.round(results.duration / 1000000); // Convert to milliseconds
    const durationText = duration > 1000 
        ? `${(duration / 1000).toFixed(1)} seconds`
        : `${duration} ms`;
    
    const sizeText = formatFileSize(results.processedSize);
    
    let html = `
        <div class="results-summary ${results.success ? 'success' : 'partial'}">
            <div class="result-item">
                <span class="label">Status:</span>
                <span class="value ${results.success ? 'success' : 'warning'}">
                    ${results.success ? '✅ All files processed successfully' : '⚠️ Some files had errors'}
                </span>
            </div>
            <div class="result-item">
                <span class="label">Total files:</span>
                <span class="value">${results.totalFiles}</span>
            </div>
            <div class="result-item">
                <span class="label">Successful:</span>
                <span class="value success">${results.successCount}</span>
            </div>
            <div class="result-item">
                <span class="label">Errors:</span>
                <span class="value ${results.errorCount > 0 ? 'error' : 'success'}">${results.errorCount}</span>
            </div>
            <div class="result-item">
                <span class="label">Processed size:</span>
                <span class="value">${sizeText}</span>
            </div>
            <div class="result-item">
                <span class="label">Duration:</span>
                <span class="value">${durationText}</span>
            </div>
        </div>
    `;
    
    if (results.errors && results.errors.length > 0) {
        html += `
            <div class="errors-section">
                <h3>❌ Errors</h3>
                <div class="error-list">
        `;
        
        results.errors.forEach((error: any) => {
            html += `
                <div class="error-item">
                    <div class="error-file">${error.filePath}</div>
                    <div class="error-message">${error.error}</div>
                </div>
            `;
        });
        
        html += `
                </div>
            </div>
        `;
    }
    
    resultsContainer.innerHTML = html;
}

function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Initialize app state
console.log('Telegram JSON to Markdown Parser loaded');
