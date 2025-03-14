package configs

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

type language struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

//go:embed language_extensions.json
var languageFile []byte

func LoadLanguageExtensions() (map[string][]string, error) {
	var languages []language
	err := json.Unmarshal(languageFile, &languages)
	if err != nil {
		return nil, fmt.Errorf("failed load language extension: %v", err)
	}
	languageMap := make(map[string][]string)
	for _, l := range languages {
		languageMap[strings.ToLower(l.Name)] = l.Extensions
	}
	return languageMap, nil
}
