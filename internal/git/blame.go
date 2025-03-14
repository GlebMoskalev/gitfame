package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type StatsEntry struct {
	Name    string
	Lines   int
	Commits int
	Files   int
}

func GetBlameStats(rs *RepositorySnapshot, useCommitter bool) ([]*StatsEntry, error) {
	regHashAndLineCommit := regexp.MustCompile(`^\S{40}\s\d+\s\d+\s\d+$`)
	regAuthor := regexp.MustCompile(`^author\s(.+)$`)
	regCommitter := regexp.MustCompile(`^committer\s(.+)$`)

	commitStatsMap := make(map[string]*StatsEntry)
	authorFilesMap := make(map[string]map[string]struct{})

	for _, file := range rs.Files {
		cmd := exec.Command("git", "blame", "--porcelain", rs.Revision, file)
		cmd.Dir = rs.GitRootDir
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git blame failed for %s: %v", file, err)
		}
		lines := strings.Split(string(out), "\n")

		for i := 0; i < len(lines); i++ {
			if regHashAndLineCommit.MatchString(lines[i]) {
				lineSplit := strings.Split(lines[i], " ")
				stats, ok := commitStatsMap[lineSplit[0]]
				if !ok {
					for j := i + 1; j < len(lines); j++ {
						if !useCommitter && regAuthor.MatchString(lines[j]) {
							lineAuthorSplit := strings.Split(lines[j], " ")
							if len(lineAuthorSplit) > 1 {
								commitStatsMap[lineSplit[0]] = &StatsEntry{
									Name: strings.Join(lineAuthorSplit[1:], " "),
								}
								stats = commitStatsMap[lineSplit[0]]
							} else {
								return nil, fmt.Errorf("failed parse line Author from %q", lines[j])
							}
							i = j
							break
						} else if useCommitter && regCommitter.MatchString(lines[j]) {
							lineCommitterSplit := strings.Split(lines[j], " ")
							if len(lineCommitterSplit) > 1 {
								commitStatsMap[lineSplit[0]] = &StatsEntry{
									Name: strings.Join(lineCommitterSplit[1:], " "),
								}
								stats = commitStatsMap[lineSplit[0]]
								i = j
								break
							} else {
								return nil, fmt.Errorf("failed parse line Committer from %q", lines[j])
							}
						}
					}
				}

				lineCount, err := strconv.Atoi(lineSplit[3])
				if err != nil {
					return nil, fmt.Errorf("failed parse line count from %q: %v", lineCount, err)
				}
				stats.Lines += lineCount

				if authorFilesMap[stats.Name] == nil {
					authorFilesMap[stats.Name] = make(map[string]struct{})
				}
				authorFilesMap[stats.Name][file] = struct{}{}
			}
		}
	}
	result := make(map[string]*StatsEntry)
	for _, s := range commitStatsMap {
		stats, ok := result[s.Name]
		if !ok {
			stats = &StatsEntry{
				Name:    s.Name,
				Lines:   0,
				Commits: 0,
			}
			result[s.Name] = stats
		}
		stats.Lines += s.Lines
		stats.Commits += 1
	}

	entries := make([]*StatsEntry, 0, len(result))
	for _, entry := range result {
		entry.Files = len(authorFilesMap[entry.Name])
		entries = append(entries, entry)
	}
	return entries, nil
}
