package stats

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/GlebMoskalev/gitfame/internal/blame"
	"github.com/GlebMoskalev/gitfame/internal/repository"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type outputFormat string

const (
	formatTabular   outputFormat = "tabular"
	formatCSV       outputFormat = "csv"
	formatJSON      outputFormat = "json"
	formatJSONLines outputFormat = "json-lines"
)

type sortField string

const (
	sortByLines   sortField = "lines"
	sortByCommits sortField = "commits"
	sortByFiles   sortField = "files"
)

func CalculateStats(
	repositoryPath, revision, extensionsArg, excludeArg, restrictArg, languagesArg, orderBy, format string,
	useCommitter bool) {
	rs, err := repository.NewRepositorySnapshot(repositoryPath, revision, extensionsArg, excludeArg, restrictArg, languagesArg)
	if err != nil {
		exitWithError("Failed to create repository snapshot", err)
	}

	contributors, err := blame.GetContributorStats(rs, useCommitter)
	if err != nil {
		exitWithError("Failed to get contributor statistics", err)
	}

	sortContributors(contributors, sortField(orderBy))

	err = outputResults(contributors, outputFormat(format), os.Stdout)
	if err != nil {
		exitWithError("Failed to output results", err)
	}

}

func exitWithError(message string, err error) {
	fmt.Printf("%s: %v\n", message, err)
	os.Exit(1)
}

func sortContributors(contributors []*blame.ContributorStats, orderBy sortField) {
	sorters := map[sortField]func(i, j int) bool{
		sortByLines: func(i, j int) bool {
			if contributors[i].Lines != contributors[j].Lines {
				return contributors[i].Lines > contributors[j].Lines
			}
			if contributors[i].Commits != contributors[j].Commits {
				return contributors[i].Commits > contributors[j].Commits
			}
			if contributors[i].Files != contributors[j].Files {
				return contributors[i].Files > contributors[j].Files
			}
			return strings.ToLower(contributors[i].Name) < strings.ToLower(contributors[j].Name)
		},
		sortByCommits: func(i, j int) bool {
			if contributors[i].Commits != contributors[j].Commits {
				return contributors[i].Commits > contributors[j].Commits
			}
			if contributors[i].Lines != contributors[j].Lines {
				return contributors[i].Lines > contributors[j].Lines
			}
			if contributors[i].Files != contributors[j].Files {
				return contributors[i].Files > contributors[j].Files
			}
			return strings.ToLower(contributors[i].Name) < strings.ToLower(contributors[j].Name)
		},
		sortByFiles: func(i, j int) bool {
			if contributors[i].Files != contributors[j].Files {
				return contributors[i].Files > contributors[j].Files
			}
			if contributors[i].Lines != contributors[j].Lines {
				return contributors[i].Lines > contributors[j].Lines
			}
			if contributors[i].Commits != contributors[j].Commits {
				return contributors[i].Commits > contributors[j].Commits
			}
			return strings.ToLower(contributors[i].Name) < strings.ToLower(contributors[j].Name)
		},
	}
	sortFunc, ok := sorters[orderBy]
	if !ok {
		exitWithError("Invalid sort field", fmt.Errorf("unknown sort field: %s", orderBy))
	}
	sort.Slice(contributors, sortFunc)
}

func outputResults(contributors []*blame.ContributorStats, format outputFormat, out io.Writer) error {
	formatters := map[outputFormat]func([]*blame.ContributorStats, io.Writer) error{
		formatTabular:   outputTabular,
		formatCSV:       outputCSV,
		formatJSON:      outputJSON,
		formatJSONLines: outputJSONLines,
	}
	formatter, ok := formatters[format]
	if !ok {
		return fmt.Errorf("unsupported format: %s", format)
	}
	return formatter(contributors, out)
}

func outputTabular(contributors []*blame.ContributorStats, out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	_, err := fmt.Fprintln(w, "Name\tLines\tCommits\tFiles")
	if err != nil {
		return fmt.Errorf("failed to output tabular: %v", err)
	}
	for _, e := range contributors {
		_, err = fmt.Fprintf(w, "%s\t%d\t%d\t%d\n", e.Name, e.Lines, e.Commits, e.Files)
		if err != nil {
			return fmt.Errorf("failed to output tabular: %v", err)
		}
	}
	err = w.Flush()
	if err != nil {
		return fmt.Errorf("failed to output tabular: %v", err)
	}
	return nil
}

func outputCSV(contributors []*blame.ContributorStats, out io.Writer) error {
	w := csv.NewWriter(out)
	if err := w.Write([]string{"Name", "Lines", "Commits", "Files"}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, e := range contributors {
		if err := w.Write([]string{e.Name, strconv.Itoa(e.Lines), strconv.Itoa(e.Commits), strconv.Itoa(e.Files)}); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	w.Flush()
	return w.Error()
}

func outputJSON(contributors []*blame.ContributorStats, out io.Writer) error {
	encoder := json.NewEncoder(out)
	return encoder.Encode(contributors)
}

func outputJSONLines(contributors []*blame.ContributorStats, out io.Writer) error {
	encoder := json.NewEncoder(out)
	for _, c := range contributors {
		if err := encoder.Encode(c); err != nil {
			return err
		}
	}
	return nil
}
