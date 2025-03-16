package blame

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/GlebMoskalev/gitfame/internal/repository"
	"github.com/GlebMoskalev/gitfame/pkg/progressbar"
)

type ContributorStats struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

var (
	regHashCommitLog     = regexp.MustCompile(`^commit\s\S{40}$`)
	regAuthorLog         = regexp.MustCompile(`^Author:\s(.*)$`)
	regHashAndLineCommit = regexp.MustCompile(`^\S{40}\s\d+\s\d+\s\d+$`)
	regAuthor            = regexp.MustCompile(`^author\s(.+)$`)
	regCommitter         = regexp.MustCompile(`^committer\s(.+)$`)
)

func GetContributorStats(rs *repository.Snapshot, useCommitter, useProgress bool) ([]*ContributorStats, error) {
	commitStatsMap := make(map[string]*ContributorStats)
	contributorFilesMap := make(map[string]map[string]struct{})

	var bar *progressbar.ProgressBar
	if useProgress {
		bar, _ = progressbar.New(len(rs.Files), os.Stdout)
	}

	for _, file := range rs.Files {
		if useProgress && bar != nil {
			bar.Tick()
		}
		if err := processFile(rs, file, commitStatsMap, contributorFilesMap, useCommitter); err != nil {
			if useProgress && bar != nil {
				bar.Tick()
			}
			return nil, err
		}
	}

	result := aggregateResults(commitStatsMap, contributorFilesMap)
	return result, nil

}

func processFile(rs *repository.Snapshot, file string, commitStatsMap map[string]*ContributorStats,
	contributorFilesMap map[string]map[string]struct{}, useCommitter bool) error {
	cmd := exec.Command("git", "blame", "--porcelain", rs.Revision, file)
	cmd.Dir = rs.GitRootDir
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git blame failed for %s: %v", file, err)
	}

	if len(out) == 0 {
		return processEmptyFile(rs, file, commitStatsMap, contributorFilesMap)
	}
	lines := strings.Split(string(out), "\n")
	return processBlameOutput(lines, commitStatsMap, contributorFilesMap, file, useCommitter)

}

func processEmptyFile(rs *repository.Snapshot, file string, commitStatsMap map[string]*ContributorStats,
	contributorFilesMap map[string]map[string]struct{}) error {
	cmd := exec.Command("git", "log", rs.Revision, "--", file)
	cmd.Dir = rs.GitRootDir
	out, err := cmd.Output()

	if err != nil {
		return fmt.Errorf("git log failed for %s: %v", file, err)
	}

	linesSplit := strings.Split(string(out), "\n")
	if len(linesSplit) < 3 || !regHashCommitLog.MatchString(linesSplit[0]) || !regAuthorLog.MatchString(linesSplit[1]) {
		return nil
	}
	commit := strings.Split(linesSplit[0], " ")[1]
	authorLine := linesSplit[1][len("Author: "):]
	author := extractName(authorLine)
	_, ok := commitStatsMap[commit]
	if !ok {
		commitStatsMap[commit] = &ContributorStats{
			Name:  author,
			Lines: 0,
		}
	}
	ensureContributorFilesMap(contributorFilesMap, author, file)
	return nil

}

func extractName(authorLine string) string {
	if idx := strings.Index(authorLine, "<"); idx > 0 {
		return strings.TrimSpace(authorLine[:idx])
	}
	return strings.TrimSpace(authorLine)
}

func processBlameOutput(lines []string, commitStatsMap map[string]*ContributorStats,
	contributorFilesMap map[string]map[string]struct{}, file string, useCommitter bool) error {
	for i := 0; i < len(lines); i++ {
		if !regHashAndLineCommit.MatchString(lines[i]) {
			continue
		}

		lineSplit := strings.Split(lines[i], " ")
		commit := lineSplit[0]
		stats, ok := commitStatsMap[commit]
		if !ok {
			stats = findAuthorInfo(lines, i, commit, commitStatsMap, useCommitter)
			if stats == nil {
				continue
			}
			i = findNextHashLine(lines, i+1) - 1
		}
		lineCount, err := strconv.Atoi(lineSplit[3])
		if err != nil {
			return fmt.Errorf("failed parse line count from %q: %v", lineCount, err)
		}
		stats.Lines += lineCount
		ensureContributorFilesMap(contributorFilesMap, stats.Name, file)
	}
	return nil

}

func findAuthorInfo(lines []string, startIdx int, commit string,
	commitStatsMap map[string]*ContributorStats, useCommitter bool) *ContributorStats {
	nameRegex := regAuthor
	if useCommitter {
		nameRegex = regCommitter
	}

	for i := startIdx; i < len(lines); i++ {
		if nameRegex.MatchString(lines[i]) {
			lineSplit := strings.Split(lines[i], " ")
			if len(lineSplit) < 1 {
				continue
			}
			name := strings.Join(lineSplit[1:], " ")
			stats := &ContributorStats{
				Name: name,
			}
			commitStatsMap[commit] = stats
			return stats
		}
	}
	return nil

}

func findNextHashLine(lines []string, startIndex int) int {
	for i := startIndex; i < len(lines); i++ {
		if regHashAndLineCommit.MatchString(lines[i]) {
			return i
		}
	}
	return len(lines)
}

func ensureContributorFilesMap(contributorFilesMap map[string]map[string]struct{}, contributor, file string) {
	if contributorFilesMap[contributor] == nil {
		contributorFilesMap[contributor] = make(map[string]struct{})
	}
	contributorFilesMap[contributor][file] = struct{}{}
}

func aggregateResults(commitStatsMap map[string]*ContributorStats,
	contributorFilesMap map[string]map[string]struct{}) []*ContributorStats {
	aggregated := make(map[string]*ContributorStats)
	for _, s := range commitStatsMap {
		stats, ok := aggregated[s.Name]
		if !ok {
			stats = &ContributorStats{
				Name:    s.Name,
				Lines:   0,
				Commits: 0,
			}
			aggregated[s.Name] = stats
		}
		stats.Lines += s.Lines
		stats.Commits += 1
	}

	result := make([]*ContributorStats, 0, len(aggregated))
	for _, entry := range aggregated {
		entry.Files = len(contributorFilesMap[entry.Name])
		result = append(result, entry)
	}
	return result

}
