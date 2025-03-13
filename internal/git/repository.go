package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetFilesRepository(path, revision, excludeArg, restrictArg string) ([]string, error) {
	if path == "" {
		path = "."
	}

	if revision == "" {
		revision = "HEAD"
	}

	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("directory does not exist: %s", path)
	}

	gitRootDir, err := getGitRootDir(path)
	if err != nil {
		return nil, err
	}

	if err = validateRevision(gitRootDir, revision); err != nil {
		return nil, err
	}

	files, err := getFilesFromGitTree(gitRootDir, revision, excludeArg, restrictArg)

	if err != nil {
		return nil, err
	}

	return files, nil
}

func getGitRootDir(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		if exitErr := err.(*exec.ExitError); exitErr.ExitCode() == 128 {
			return "", fmt.Errorf("not a git repository: %s:", path)
		}
		return "", fmt.Errorf("failed to get git root: %v", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func validateRevision(gitRootDir, revision string) error {
	cmd := exec.Command("git", "cat-file", "commit", revision)
	cmd.Dir = gitRootDir
	if err := cmd.Run(); err != nil {
		if exitErr := err.(*exec.ExitError); exitErr.ExitCode() == 128 {
			return fmt.Errorf("invalid revision: %s", revision)
		}
		return fmt.Errorf("failed to validate revision: %v", err)
	}
	return nil
}

func getFilesFromGitTree(gitRootDir, revisionTree, excludeArg, restrictArg string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", revisionTree)
	cmd.Dir = gitRootDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	excludePatters, restrictPatterns := make([]string, 0), make([]string, 0)
	if len(excludePatters) != 0 {
		excludePatters = strings.Split(excludeArg, ",")
	}
	if len(restrictPatterns) != 0 {
		restrictPatterns = strings.Split(restrictArg, ",")
	}

	var files []string
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[1] == "blob" {
			file := fields[3]
			if len(restrictPatterns) != 0 {
				if matchesAnyPattern(file, restrictPatterns) {
					files = append(files, file)
				}
			} else if len(excludePatters) != 0 {
				if !matchesAnyPattern(file, excludePatters) {
					files = append(files, file)
				} else {
					fmt.Println(file)
				}
			} else {
				files = append(files, file)
			}

		}

	}
	return files, nil
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
