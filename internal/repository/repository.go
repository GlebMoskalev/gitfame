package repository

import (
	"fmt"
	"github.com/GlebMoskalev/gitfame/configs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type Snapshot struct {
	GitRootDir string
	Files      []string
	Revision   string
	Filters    Filters
}
type Filters struct {
	ExcludePatterns  []string
	RestrictPatterns []string
	Extensions       []string
}

func NewRepositorySnapshot(
	repositoryPath,
	revision,
	extensionsArg,
	excludeArg,
	restrictArg,
	languagesArg string) (*Snapshot, error) {
	rs := &Snapshot{
		Revision: revision,
		Filters:  createFilters(extensionsArg, excludeArg, restrictArg, languagesArg),
	}
	if err := rs.getGitRootDir(repositoryPath); err != nil {
		return nil, err
	}

	if err := rs.validateRevision(); err != nil {
		return nil, err
	}

	if err := rs.getFilesFromGitTree(); err != nil {
		return nil, err
	}

	return rs, nil

}

func createFilters(extensionsArg, excludeArg, restrictArg, languagesArg string) Filters {
	filters := Filters{
		ExcludePatterns:  splitIfNotEmpty(excludeArg),
		RestrictPatterns: splitIfNotEmpty(restrictArg),
		Extensions:       splitIfNotEmpty(extensionsArg),
	}
	if languagesArg != "" {
		languagesMap, err := configs.LoadLanguageExtensions()
		if err == nil {
			languages := strings.Split(languagesArg, ",")
			for _, l := range languages {
				languageExtensions, ok := languagesMap[l]
				if !ok {
					continue
				}
				filters.Extensions = append(filters.Extensions, languageExtensions...)
			}
		}
	}
	return filters
}

func splitIfNotEmpty(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func (rs *Snapshot) getGitRootDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		if exitErr := err.(*exec.ExitError); exitErr.ExitCode() == 128 {
			return fmt.Errorf("not a git repository: %s", path)
		}
		return fmt.Errorf("failed to get git root: %v", err)
	}
	rs.GitRootDir = strings.TrimSpace(string(out))
	return nil

}

func (rs *Snapshot) validateRevision() error {
	cmd := exec.Command("git", "cat-file", "commit", rs.Revision)
	cmd.Dir = rs.GitRootDir
	if err := cmd.Run(); err != nil {
		if exitErr := err.(*exec.ExitError); exitErr.ExitCode() == 128 {
			return fmt.Errorf("invalid revision: %s", rs.Revision)
		}
		return fmt.Errorf("failed to validate revision: %v", err)
	}

	return nil

}

func (rs *Snapshot) getFilesFromGitTree() error {
	cmd := exec.Command("git", "ls-tree", "-r", rs.Revision)
	cmd.Dir = rs.GitRootDir
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var files []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		if fields[1] == "blob" {
			file := strings.Join(fields[3:], " ")
			addFile := false
			if len(rs.Filters.RestrictPatterns) != 0 {
				if matchesAnyPattern(file, rs.Filters.RestrictPatterns) {
					addFile = true
				}
			} else {
				addFile = true
			}
			if len(rs.Filters.ExcludePatterns) != 0 {
				if matchesAnyPattern(file, rs.Filters.ExcludePatterns) {
					addFile = false
				}
			}

			if addFile && len(rs.Filters.Extensions) != 0 {
				if slices.Contains(rs.Filters.Extensions, filepath.Ext(file)) {
					files = append(files, file)
				}
			} else if addFile {
				files = append(files, file)
			}
		}

	}
	rs.Files = files
	return nil

}

func (rs *Snapshot) filterRepoFiles(lines []string) []string {
	var files []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 4 || fields[1] != "blob" {
			continue
		}
		file := strings.Join(fields[3:], " ")
		if rs.shouldIncludeFile(file) {
			files = append(files, file)
		}
	}
	return files
}

func (rs *Snapshot) shouldIncludeFile(file string) bool {
	if len(rs.Filters.RestrictPatterns) > 0 {
		if !matchesAnyPattern(file, rs.Filters.ExcludePatterns) {
			return false
		}
	}

	if len(rs.Filters.ExcludePatterns) > 0 && matchesAnyPattern(file, rs.Filters.ExcludePatterns) {
		return false
	}

	if len(rs.Filters.Extensions) > 0 {
		return slices.Contains(rs.Filters.Extensions, filepath.Ext(file))
	}
	return true

}

func matchesAnyPattern(file string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		matched, _ := filepath.Match(pattern, file)
		if matched {
			return true
		}
	}
	return false
}
