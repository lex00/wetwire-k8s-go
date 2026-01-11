package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/lex00/wetwire-k8s-go/internal/build"
	"github.com/urfave/cli/v2"
)

// watchCommand creates the watch subcommand
func watchCommand() *cli.Command {
	return &cli.Command{
		Name:      "watch",
		Usage:     "Monitor source files and auto-rebuild on changes",
		ArgsUsage: "[PATH]",
		Description: `Watch monitors Go source files for changes and automatically
rebuilds Kubernetes manifests when changes are detected.

If PATH is not specified, the current directory is used.

The watcher uses debouncing to avoid multiple rapid rebuilds when
several files change in quick succession.

Examples:
  wetwire-k8s watch                          # Watch current directory
  wetwire-k8s watch ./k8s                    # Watch specific directory
  wetwire-k8s watch -o manifests.yaml ./k8s  # Output to specific file
  wetwire-k8s watch --interval 500ms ./k8s   # Custom debounce interval`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file path (use '-' for stdout)",
				Value:   "-",
			},
			&cli.DurationFlag{
				Name:  "interval",
				Usage: "Debounce interval for file changes",
				Value: 300 * time.Millisecond,
			},
		},
		Action: runWatch,
	}
}

// runWatch executes the watch command
func runWatch(c *cli.Context) error {
	// Determine source path
	sourcePath := c.Args().First()
	if sourcePath == "" {
		sourcePath = "."
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Validate source path exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("source path does not exist: %s", absPath)
	}

	outputPath := c.String("output")
	interval := c.Duration("interval")

	// Get output writer
	writer := c.App.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Run watch loop (blocks until interrupted)
	return runWatchLoop(absPath, outputPath, interval, writer)
}

// runWatchLoop runs the main watch loop
func runWatchLoop(sourcePath, outputPath string, interval time.Duration, writer io.Writer) error {
	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Add source path to watcher
	if err := addWatchPath(watcher, sourcePath); err != nil {
		return fmt.Errorf("failed to add watch path: %w", err)
	}

	// Perform initial build
	fmt.Fprintln(writer, "Starting watch mode...")
	if err := performBuild(sourcePath, outputPath, writer); err != nil {
		fmt.Fprintf(writer, "Initial build failed: %v\n", err)
	}

	// Debounce timer
	var debounceTimer *time.Timer
	var debounceMutex sync.Mutex

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to Go file changes
			if !isGoFile(event.Name) {
				continue
			}

			// Check for write or create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce: reset timer on each event
			debounceMutex.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(interval, func() {
				fmt.Fprintf(writer, "\nFile changed: %s\n", filepath.Base(event.Name))
				if err := performBuild(sourcePath, outputPath, writer); err != nil {
					fmt.Fprintf(writer, "Build failed: %v\n", err)
				} else {
					fmt.Fprintln(writer, "Build complete")
				}
			})
			debounceMutex.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(writer, "Watcher error: %v\n", err)
		}
	}
}

// runWatchWithContext runs watch with a context for cancellation (used in tests)
func runWatchWithContext(ctx context.Context, sourcePath, outputPath string, interval time.Duration, writer io.Writer) error {
	return runWatchWithContextAndCallback(ctx, sourcePath, outputPath, interval, writer, nil)
}

// runWatchWithContextAndCallback runs watch with context and build callback (used in tests)
func runWatchWithContextAndCallback(ctx context.Context, sourcePath, outputPath string, interval time.Duration, writer io.Writer, onBuild func()) error {
	// Resolve to absolute path
	absPath, err := filepath.Abs(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	// Add source path to watcher
	if err := addWatchPath(watcher, absPath); err != nil {
		return fmt.Errorf("failed to add watch path: %w", err)
	}

	// Perform initial build
	if err := performBuild(absPath, outputPath, writer); err != nil {
		fmt.Fprintf(writer, "Initial build failed: %v\n", err)
	}
	if onBuild != nil {
		onBuild()
	}

	// Debounce timer
	var debounceTimer *time.Timer
	var debounceMutex sync.Mutex

	for {
		select {
		case <-ctx.Done():
			debounceMutex.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceMutex.Unlock()
			return ctx.Err()

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Only react to Go file changes
			if !isGoFile(event.Name) {
				continue
			}

			// Check for write or create events
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce: reset timer on each event
			debounceMutex.Lock()
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(interval, func() {
				if err := performBuild(absPath, outputPath, writer); err != nil {
					fmt.Fprintf(writer, "Build failed: %v\n", err)
				}
				if onBuild != nil {
					onBuild()
				}
			})
			debounceMutex.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(writer, "Watcher error: %v\n", err)
		}
	}
}

// addWatchPath adds a path (and subdirectories) to the watcher
func addWatchPath(watcher *fsnotify.Watcher, path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return watcher.Add(walkPath)
		}
		return nil
	})
}

// isGoFile checks if a file has a .go extension
func isGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

// performBuild runs the build pipeline and writes output
func performBuild(sourcePath, outputPath string, writer io.Writer) error {
	// Run the build pipeline
	result, err := build.Build(sourcePath, build.Options{
		OutputMode: build.SingleFile,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// No resources found
	if len(result.OrderedResources) == 0 {
		return nil
	}

	// Generate YAML output
	output, err := generateOutput(result.OrderedResources, "yaml")
	if err != nil {
		return fmt.Errorf("failed to generate output: %w", err)
	}

	// Write output
	if outputPath == "-" {
		// Write to stdout/writer
		_, err = writer.Write(output)
		return err
	}

	// Write to file
	if err := os.WriteFile(outputPath, output, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
