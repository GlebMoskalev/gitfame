package integration

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const (
	mainProjectPath = "github.com/GlebMoskalev/gitfame/cmd/gitfame"
	nameBinary      = "gitfametest"
)

type TestCase struct {
	*TestDescription
	Expected []byte
}

type TestDescription struct {
	Name   string   `yaml:"name"`
	Args   []string `yaml:"args"`
	Bundle string   `yaml:"bundle"`
	Error  bool     `yaml:"error,omitempty"`
	Format string   `yaml:"format,omitempty"`
}

func TestGitFame(t *testing.T) {
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, nameBinary)
	err := buildBinary(binaryPath)
	if err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	bundleNameToPath := getBundleNamesToPath(t, tempDir)
	for _, ts := range readTestCases(t) {
		t.Run(ts.Name, func(t *testing.T) {
			bundlePath, ok := bundleNameToPath[ts.Bundle]
			if !ok {
				t.Fatalf("failed to get bundle %q", ts.Bundle)
			}
			args := []string{"--repository", bundlePath}
			args = append(args, ts.Args...)
			cmd := exec.Command(fmt.Sprintf("./%s", nameBinary), args...)
			cmd.Dir = tempDir
			output, err := cmd.Output()
			if !ts.Error {
				assert.NoError(t, err)
				equalResult(t, ts.Expected, output, ts.Format)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func equalResult(t *testing.T, expected, actual []byte, format string) {
	t.Helper()

	switch format {
	case "json":
		fmt.Println()
		assert.JSONEq(t, string(expected), string(actual))
	case "json-lines":
		expectedLines := bytes.Split(bytes.TrimSpace(expected), []byte("\n"))
		actualLines := bytes.Split(bytes.TrimSpace(actual), []byte("\n"))
		assert.Equal(t, len(expectedLines), len(actualLines))
		for i, l := range expectedLines {
			assert.JSONEq(t, string(l), string(actualLines[i]))
		}
	default:
		assert.Equal(t, string(expected), string(actual))
	}
}

func readTestCases(t *testing.T) []*TestCase {
	testsDir := "./testdata/tests"
	entries, err := os.ReadDir(testsDir)
	if err != nil {
		t.Fatalf("failed read testsCases directory %q: %v", testsDir, err)
	}
	testCases := make([]*TestCase, 0)
	for _, e := range entries {
		if e.IsDir() {
			testCase := readTestCase(t, filepath.Join(testsDir, e.Name()))
			testCase.Name = fmt.Sprintf("%s %s", e.Name(), testCase.Name)
			testCases = append(testCases, testCase)
		}
	}
	return testCases
}

func readTestCase(t *testing.T, path string) *TestCase {
	desc := readTestDescription(t, path)
	expected, err := os.ReadFile(filepath.Join(path, "expected.out"))
	if err != nil {
		t.Fatalf("failed to read expected.out at %q: %v", path, err)
	}
	return &TestCase{
		TestDescription: desc,
		Expected:        expected,
	}
}

func readTestDescription(t *testing.T, path string) *TestDescription {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(path, "description.yaml"))
	if err != nil {
		t.Fatalf("failed to read description.yaml at %q: %v", path, err)
	}
	assert.NoError(t, err)
	var testDes TestDescription
	if err := yaml.Unmarshal(data, &testDes); err != nil {
		t.Fatalf("failed to unmarshal YAML at %q: %v", path, err)
	}
	return &testDes
}

func buildBinary(binaryPath string) error {
	cmd := exec.Command("go", "build", "-o", binaryPath, mainProjectPath)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getBundleNamesToPath(t *testing.T, tempDir string) map[string]string {
	bundlesDir := filepath.Join("./testdata", "bundles")
	entries, err := os.ReadDir(bundlesDir)
	if err != nil {
		t.Fatalf("failed read bundles directory: %v", err)
	}

	bundleNameToPath := make(map[string]string)
	for _, e := range entries {
		destinationBundlePath := filepath.Join(tempDir, e.Name())
		gitCloneBundle(t, filepath.Join(bundlesDir, e.Name()), destinationBundlePath)
		bundleNameToPath[e.Name()] = destinationBundlePath
	}

	return bundleNameToPath
}

func gitCloneBundle(t *testing.T, sourcePath, destinationPath string) {
	t.Helper()
	cmd := exec.Command("git", "clone", sourcePath, destinationPath)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to git clone %q to %q: %v", sourcePath, destinationPath, err)
	}
}
