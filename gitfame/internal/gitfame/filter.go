package gitfame

import (
	"log"
	"path/filepath"
	"strings"
)

func filter(filePaths []string) []string {
	filePaths = filterExtensions(filePaths)
	filePaths = filterLanguages(filePaths)
	filePaths = filterExclude(filePaths)
	filePaths = filterRestrictTo(filePaths)
	return filePaths
}

func filterExtensions(filePaths []string) []string {
	if *flags.extensions == "" {
		return filePaths
	}
	filtred := make([]string, 0)
	extensions := strings.Split(*flags.extensions, ",")
	for _, file := range filePaths {
		for _, extension := range extensions {
			if strings.HasSuffix(file, extension) {
				filtred = append(filtred, file)
			}
		}
	}
	return filtred
}

func filterLanguages(filePaths []string) []string {
	if *flags.languages == "" {
		return filePaths
	}

	LanguagesMap := initLanguages()
	extensions := make([]string, 0)
	filtred := make([]string, 0)

	languages := strings.Split(*flags.languages, ",")
	for _, language := range languages {
		extensions = append(extensions, LanguagesMap[strings.ToLower(language)]...)
	}

	for _, file := range filePaths {
		for _, extension := range extensions {
			if strings.HasSuffix(file, extension) {
				filtred = append(filtred, file)
			}
		}
	}

	return filtred
}

func filterExclude(filePaths []string) []string {
	if *flags.exclude == "" {
		return filePaths
	}

	filtred := make([]string, 0)
	patterns := strings.Split(*flags.exclude, ",")
loop:
	for _, file := range filePaths {
		for _, pattern := range patterns {
			match, err := filepath.Match(pattern, file)
			if err != nil {
				log.Fatal(err)
			}
			if match {
				continue loop
			}
		}
		filtred = append(filtred, file)
	}

	return filtred
}

func filterRestrictTo(filePaths []string) []string {
	if *flags.restrictTo == "" {
		return filePaths
	}

	filtred := make([]string, 0)
	patterns := strings.Split(*flags.restrictTo, ",")
loop:
	for _, file := range filePaths {
		for _, pattern := range patterns {
			match, err := filepath.Match(pattern, file)
			if err != nil {
				log.Fatal(err)
			}
			if match {
				filtred = append(filtred, file)
				continue loop
			}
		}
	}

	return filtred
}
