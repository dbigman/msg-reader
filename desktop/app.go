package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// OpenFile allows the frontend to request opening a file
func (a *App) OpenFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// SaveFile allows the frontend to save a file
func (a *App) SaveFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return nil
}

// OpenFileDialog opens a file dialog and returns the selected file paths
func (a *App) OpenFileDialog() ([]string, error) {
	files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select MSG or EML files",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Email Files (*.msg, *.eml)",
				Pattern:     "*.msg;*.eml",
			},
			{
				DisplayName: "MSG Files (*.msg)",
				Pattern:     "*.msg",
			},
			{
				DisplayName: "EML Files (*.eml)",
				Pattern:     "*.eml",
			},
			{
				DisplayName: "All Files (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	
	return files, err
}

// SaveFileDialog opens a save file dialog and returns the selected file path
func (a *App) SaveFileDialog(defaultFilename string) (string, error) {
	file, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save File",
		DefaultFilename: defaultFilename,
	})
	
	return file, err
} 