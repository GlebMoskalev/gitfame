package git

import (
	"fmt"
	"github.com/GlebMoskalev/gitfame/configs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type RepositorySnapshot struct {
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
	languagesArg string) (*RepositorySnapshot, error) {
	rs := &RepositorySnapshot{
		Revision: revision,
	}
	if err := rs.getGitRootDir(repositoryPath); err != nil {
		return nil, err
	}
	if err := rs.validateRevision(); err != nil {
		return nil, err
	}
	excludePatters, restrictPatterns, extensions := make([]string, 0), make([]string, 0), make([]string, 0)
	if excludeArg != "" {
		excludePatters = strings.Split(excludeArg, ",")
	}
	if restrictArg != "" {
		restrictPatterns = strings.Split(restrictArg, ",")
	}
	if extensionsArg != "" {
		extensions = strings.Split(extensionsArg, ",")
	}
	rs.Filters.ExcludePatterns = excludePatters
	rs.Filters.RestrictPatterns = restrictPatterns
	rs.Filters.Extensions = extensions

	if languagesArg != "" {
		languagesMap, err := configs.LoadLanguageExtensions()
		if err != nil {
			return nil, err
		}
		languages := strings.Split(languagesArg, ",")
		for _, l := range languages {
			languageExtensions, ok := languagesMap[l]
			if !ok {
				continue
				//return nil, fmt.Errorf("language %q is not in the supported list (available: %v)", l, maps.Keys(languagesMap))
			}
			rs.Filters.Extensions = append(rs.Filters.Extensions, languageExtensions...)
		}
	}

	if err := rs.getFilesFromGitTree(); err != nil {
		return nil, err
	}

	return rs, nil
}

func (rs *RepositorySnapshot) getGitRootDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("directory does not exist: %s", path)
	}

	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		if exitErr := err.(*exec.ExitError); exitErr.ExitCode() == 128 {
			return fmt.Errorf("not a git repository: %s:", path)
		}
		return fmt.Errorf("failed to get git root: %v", err)
	}
	rs.GitRootDir = strings.TrimSpace(string(out))
	return nil
}

func (rs *RepositorySnapshot) validateRevision() error {
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

func (rs *RepositorySnapshot) getFilesFromGitTree() error {
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
			file := fields[3]
			addFile := false
			if len(rs.Filters.RestrictPatterns) != 0 {
				if matchesAnyPattern(file, rs.Filters.RestrictPatterns) {
					addFile = true
				}
			} else if len(rs.Filters.ExcludePatterns) != 0 {
				if !matchesAnyPattern(file, rs.Filters.ExcludePatterns) {
					addFile = true
				}
			} else {
				addFile = true
			}

			if addFile && len(rs.Filters.Extensions) != 0 {
				if slices.Contains(rs.Filters.Extensions, filepath.Ext(file)) {
					files = append(files, file)
				}
			} else {
				files = append(files, file)
			}
		}

	}
	rs.Files = files
	return nil
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
