//go:build !solution

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab.com/slon/shad-go/gitfame/internal/gitfame"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Parse command line arguments
	var (
		repository   string
		revision     string
		orderBy      string
		useCommitter bool
		format       string
		extensions   string
		languages    string
		exclude      string
		restrictTo   string
		progress     bool
	)

	flag.StringVar(&repository, "repository", ".", "path to Git repository")
	flag.StringVar(&revision, "revision", "HEAD", "Git revision to analyze")
	flag.StringVar(&orderBy, "order-by", "lines", "sort order: lines, commits, or files")
	flag.BoolVar(&useCommitter, "use-committer", false, "use committer instead of author")
	flag.StringVar(&format, "format", "tabular", "output format: tabular, csv, json, or json-lines")
	flag.StringVar(&extensions, "extensions", "", "file extensions to include (comma-separated)")
	flag.StringVar(&languages, "languages", "", "languages to include (comma-separated)")
	flag.StringVar(&exclude, "exclude", "", "glob patterns to exclude (comma-separated)")
	flag.StringVar(&restrictTo, "restrict-to", "", "glob patterns to restrict to (comma-separated)")
	flag.BoolVar(&progress, "progress", false, "show progress")

	flag.Parse()

	// Create and validate configuration
	cfg := &gitfame.Config{
		Repository:   repository,
		Revision:     revision,
		OrderBy:      orderBy,
		UseCommitter: useCommitter,
		Format:       format,
		Progress:     progress,
	}

	// Parse extensions
	if extensions != "" {
		cfg.Extensions = strings.Split(extensions, ",")
	}

	// Parse languages
	if languages != "" {
		cfg.Languages = strings.Split(languages, ",")
		// Convert languages to lowercase
		for i, lang := range cfg.Languages {
			cfg.Languages[i] = strings.ToLower(strings.TrimSpace(lang))
		}
	}

	// Parse exclude patterns
	if exclude != "" {
		cfg.Exclude = strings.Split(exclude, ",")
	}

	// Parse restrict-to patterns
	if restrictTo != "" {
		cfg.RestrictTo = strings.Split(restrictTo, ",")
	}

	// Validate configuration
	if err := gitfame.ValidateConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load language extension mappings
	configDir := filepath.Dir(filepath.Dir(filepath.Dir(os.Args[0])))
	languageConfigPath := filepath.Join(configDir, "configs", "language_extensions.json")

	// If the config is not found at the relative path, try to find it using an absolute path
	// from the current working directory
	if _, err := os.Stat(languageConfigPath); os.IsNotExist(err) {
		execPath, err := os.Executable()
		if err == nil {
			execDir := filepath.Dir(execPath)
			rootDir := filepath.Dir(filepath.Dir(execDir))
			languageConfigPath = filepath.Join(rootDir, "configs", "language_extensions.json")
		}
	}

	// If still not found, try relative to the current working directory
	if _, err := os.Stat(languageConfigPath); os.IsNotExist(err) {
		cwd, err := os.Getwd()
		if err == nil {
			for i := 0; i < 5; i++ { // Try up to 5 levels up
				testPath := filepath.Join(cwd, "gitfame", "configs", "language_extensions.json")
				if _, err := os.Stat(testPath); err == nil {
					languageConfigPath = testPath
					break
				}
				cwd = filepath.Dir(cwd)
			}
		}
	}

	var extensionMap map[string][]string
	var err error

	if cfg.Languages != nil && len(cfg.Languages) > 0 {
		extensionMap, err = gitfame.LoadLanguageExtensions(languageConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load language extensions: %v\n", err)
			fmt.Fprintf(os.Stderr, "Language filtering will be disabled\n")
		}
	}

	// Process repository with timeout
	processDone := make(chan struct{})
	var stats []*gitfame.AuthorStat
	var processErr error

	go func() {
		stats, processErr = gitfame.ProcessRepo(ctx, cfg, extensionMap)
		close(processDone)
	}()

	// Wait for processing to complete with a timeout
	select {
	case <-processDone:
		// Processing completed
	case <-time.After(30 * time.Minute): // Overall timeout for long-running processes
		cancel()
		fmt.Fprintf(os.Stderr, "Error: processing timed out after 30 minutes\n")
		os.Exit(1)
	}

	if processErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", processErr)
		os.Exit(1)
	}

	// Format and output results
	if err := gitfame.FormatOutput(stats, cfg.Format, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
