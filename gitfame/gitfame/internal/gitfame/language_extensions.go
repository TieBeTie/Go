package gitfame

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadLanguageExtensions loads language-to-extension mappings from a JSON file
func LoadLanguageExtensions(jsonPath string) (map[string][]string, error) {
	// If the path is empty, try to find the language_extensions.json file
	if jsonPath == "" {
		// Try to locate the file in several common locations
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}

		// Define potential locations to check
		possiblePaths := []string{
			filepath.Join(wd, "language_extensions.json"),
			filepath.Join(wd, "gitfame", "language_extensions.json"),
			filepath.Join(wd, "internal", "gitfame", "language_extensions.json"),
			filepath.Join(wd, "cmd", "gitfame", "language_extensions.json"),
			filepath.Join(wd, "data", "language_extensions.json"),
			"/etc/gitfame/language_extensions.json",
			"/usr/local/share/gitfame/language_extensions.json",
		}

		// Try each location
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				jsonPath = path
				fmt.Fprintf(os.Stderr, "DEBUG: Found language extensions file at %s\n", jsonPath)
				break
			}
		}

		// If we still don't have a path, return an empty map
		if jsonPath == "" {
			fmt.Fprintf(os.Stderr, "WARNING: No language extensions file found, language filtering will not work\n")
			return make(map[string][]string), nil
		}
	}

	// Read and parse the JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "WARNING: Language extensions file %s not found, language filtering will not work\n", jsonPath)
			return make(map[string][]string), nil
		}
		return nil, fmt.Errorf("failed to read language extensions file: %w", err)
	}

	var extensionMap map[string][]string
	if err := json.Unmarshal(data, &extensionMap); err != nil {
		return nil, fmt.Errorf("failed to parse language extensions file: %w", err)
	}

	// Add debug info
	fmt.Fprintf(os.Stderr, "DEBUG: Loaded %d language mappings from %s\n", len(extensionMap), jsonPath)
	return extensionMap, nil
}
