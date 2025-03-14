package git

import (
	"fmt"
	"github.com/GlebMoskalev/gitfame/configs"
)

type Stats struct {
	Name    string
	Lines   int
	Commits int
	Files   int
}

func CalculateStats(repositoryPath, revision, extensionsArg, excludeArg, restrictArg, languagesArg string, useCommitter bool) {
	configs.LoadLanguageExtensions()
	rs, _ := NewRepositorySnapshot(repositoryPath, revision, extensionsArg, excludeArg, restrictArg, languagesArg)
	blameEntries, _ := GetBlameStats(rs, useCommitter)
	for _, entry := range blameEntries {
		fmt.Println(entry)
	}
}
