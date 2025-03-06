package main

import (
	"embed"
	"log"
	"os"
	"fmt"
	"strings"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	// Debug: Print executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
	} else {
		fmt.Println("Executable path:", execPath)
		fmt.Println("Working directory:", filepath.Dir(execPath))
	}
	
	// Debug: Print current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
	} else {
		fmt.Println("Current working directory:", cwd)
	}
	
	// Create an instance of the app structure
	app := NewApp()

	// Get command line arguments (files to open)
	args := os.Args[1:]
	
	// Debug: Print command line arguments
	fmt.Println("Command line arguments:", args)
	
	// Filter out macOS specific flags
	var filesToOpen []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// Skip macOS specific flags and their values
		if strings.HasPrefix(arg, "-NS") || strings.HasPrefix(arg, "-Apple") || strings.HasPrefix(arg, "-psn_") {
			// Skip the flag and its value if it has one
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		}
		
		// Check if the argument is a file that exists
		if fileInfo, err := os.Stat(arg); err == nil && !fileInfo.IsDir() {
			// Get the absolute path to avoid any path resolution issues
			absPath, err := filepath.Abs(arg)
			if err == nil {
				filesToOpen = append(filesToOpen, absPath)
				fmt.Println("Added file to open:", absPath)
			} else {
				fmt.Println("Error getting absolute path for", arg, ":", err)
				filesToOpen = append(filesToOpen, arg)
			}
		} else {
			fmt.Println("Argument is not a valid file or is a directory:", arg)
			if err != nil {
				fmt.Println("Error:", err)
			}
			
			// Try to resolve the path relative to the current working directory
			relPath := filepath.Join(cwd, arg)
			fmt.Println("Trying relative path:", relPath)
			if fileInfo, err := os.Stat(relPath); err == nil && !fileInfo.IsDir() {
				absPath, err := filepath.Abs(relPath)
				if err == nil {
					filesToOpen = append(filesToOpen, absPath)
					fmt.Println("Added file to open (relative path):", absPath)
				} else {
					fmt.Println("Error getting absolute path for", relPath, ":", err)
					filesToOpen = append(filesToOpen, relPath)
				}
			} else {
				fmt.Println("Relative path is not a valid file or is a directory:", relPath)
				if err != nil {
					fmt.Println("Error:", err)
				}
			}
		}
	}
	
	fmt.Println("Files to open:", filesToOpen)
	
	if len(filesToOpen) > 0 {
		// Store the files to open on startup
		app.filesToOpenOnStartup = filesToOpen
		fmt.Println("Files to open on startup:", app.filesToOpenOnStartup)
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "msgReader",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
} 