package git

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/GlebMoskalev/gitfame/configs"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
)

func CalculateStats(repositoryPath,
	revision,
	extensionsArg,
	excludeArg,
	restrictArg,
	languagesArg,
	orderBy,
	format string,
	useCommitter bool) {
	configs.LoadLanguageExtensions()
	rs, err := NewRepositorySnapshot(repositoryPath, revision, extensionsArg, excludeArg, restrictArg, languagesArg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	blameEntries, err := GetBlameStats(rs, useCommitter)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var funcSort func(i, j int) bool
	switch orderBy {
	case "lines":
		funcSort = func(i, j int) bool {
			return blameEntries[i].Lines > blameEntries[j].Lines
		}
	case "commits":
		funcSort = func(i, j int) bool {
			return blameEntries[i].Commits > blameEntries[j].Commits
		}
	case "files":
		funcSort = func(i, j int) bool {
			return blameEntries[i].Commits > blameEntries[j].Commits
		}
	default:
		fmt.Println("error")
	}
	sort.Slice(blameEntries, funcSort)

	switch format {
	case "tabular":
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")
		for _, e := range blameEntries {
			fmt.Fprintln(w, e.Name, "\t", e.Lines, "\t", e.Commits, "\t", e.Files)
		}
		w.Flush()
	case "csv":
		w := csv.NewWriter(os.Stdout)
		if err = w.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, e := range blameEntries {
			if err := w.Write([]string{e.Name, strconv.Itoa(e.Files), strconv.Itoa(e.Commits), strconv.Itoa(e.Files)}); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		w.Flush()
	case "json":
		b, err := json.Marshal(blameEntries)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(b))
	case "json-lines":
		for _, e := range blameEntries {
			b, err := json.Marshal(e)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(string(b))
		}
	default:
		fmt.Println("error")
	}
}
