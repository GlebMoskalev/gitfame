package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type BlameEntry struct {
	Name  string
	Lines int
	Files int
}

func GetBlameStats(rs *RepositorySnapshot, useCommitter bool) ([]*BlameEntry, error) {
	type blameStats struct {
		name  string
		lines int
		files map[string]struct{}
	}

	regHashAndLineCommit := regexp.MustCompile(`^\S{40}\s\d+\s\d+\s\d+$`)
	regAuthor := regexp.MustCompile(`^author\s(.+)$`)
	regCommitter := regexp.MustCompile(`^committer\s(.+)$`)

	commitStatsMap := make(map[string]*blameStats)
	currentCommitHash := ""

	for _, file := range rs.Files {
		cmd := exec.Command("git", "blame", "--porcelain", rs.Revision, file)
		cmd.Dir = rs.GitRootDir
		out, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git blame failed for %s: %v", file, err)
		}

		for _, line := range bytes.Split(out, []byte("\n")) {
			lineStr := string(line)
			if regHashAndLineCommit.MatchString(lineStr) {
				lineSplit := strings.Split(lineStr, " ")
				if len(lineSplit) < 4 {
					continue
				}
				currentCommitHash = lineSplit[0]
				bs, ok := commitStatsMap[currentCommitHash]
				lineInt, err := strconv.Atoi(lineSplit[3])
				if err != nil {
					return nil, fmt.Errorf("failed parse line count from %q: %v", lineInt, err)
				}
				if !ok {
					bs = &blameStats{
						lines: 0,
						files: make(map[string]struct{}),
					}
					commitStatsMap[currentCommitHash] = bs
				}
				bs.lines += lineInt
				bs.files[file] = struct{}{}
			} else if regAuthor.MatchString(lineStr) && !useCommitter {
				lineSplit := strings.Split(lineStr, " ")
				if len(lineSplit) > 1 {
					if bs, ok := commitStatsMap[currentCommitHash]; ok {
						bs.name = lineSplit[1]
					}
				}

			} else if regCommitter.MatchString(lineStr) {
				lineSplit := strings.Split(lineStr, " ")
				if len(lineSplit) > 1 {
					if bs, ok := commitStatsMap[currentCommitHash]; ok {
						bs.name = lineSplit[1]
					}
				}
			}
		}
	}

	result := make(map[string]*BlameEntry)
	for _, b := range commitStatsMap {
		if _, ok := result[b.name]; !ok {
			result[b.name] = &BlameEntry{
				Name: b.name,
			}
		}
		result[b.name].Lines += b.lines
		result[b.name].Files += len(b.files)
	}

	entries := make([]*BlameEntry, 0, len(result))
	for _, entry := range result {
		entries = append(entries, entry)
	}
	return entries, nil
}
