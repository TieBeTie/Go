package gitfame

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type language struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Extensions []string `json:"extensions"`
}

func initLanguages() map[string][]string {
	LanguagesMap := make(map[string][]string)

	data, err := os.ReadFile("../../configs/language_extensions.json")
	if err != nil {
		log.Fatal(err)
	}

	var languages []language
	if err := json.Unmarshal(data, &languages); err != nil {
		log.Fatal(err)
	}

	for _, lg := range languages {
		LanguagesMap[strings.ToLower(lg.Name)] = lg.Extensions
	}
	return LanguagesMap
}
