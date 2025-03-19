# GitFame

A console utility for calculating author statistics in Git repositories.

## Description

GitFame analyzes a Git repository and provides statistics about authors (or committers) based on a specific commit.The utility calculates:
- Number of lines of code
- Number of commits
- Number of modified files

Example output:
```
 ./gitfame --repository=. --extensions=".go"
Name             Lines  Commits  Files  
GlebMoskalev     977    14       7  
AlexDeveloper    642    10       5  
DianaMaintainer  315    6        3  
BobContributor   188    4        2  
CharlieDevOps    120    3        1 
```

## Installation
1. Clone the repository:
```bash
git clone https://github.com/GlebMoskalev/gitfame
```
2. Navigate to the project directory:
```bash
cd cmd/gitfame
```
3. Build the project:
 ```bash
go build
```

## Usage
Basic syntax:
```
gitfame [flags]
```
### Available Flags
- ``--repository`` – Path to the Git repository (default: current directory .)
- ``--revision`` – Commit reference (default: HEAD)
- ``--revision`` – Commit reference (default: HEAD)
- ``--order-by`` — Sort results by: lines (default), commits, files
- ``--use-committer`` — Use committer instead of author (default: false)
- ``--format`` — Output format: tabular (default), csv, json, json-lines
- ``--extensions`` — Filter by file extensions (e.g., .go,.md)
- ``--extensions`` — Filter by file extensions (e.g., .go,.md)
- ``--languages`` — Filter by languages (e.g., go,markdown)
- ``--exclude`` — Exclude files matching Glob patterns (e.g., foo/*,bar/*)
- ``--restrict-to`` — Restrict analysis to files matching Glob patterns
- ``--progress`` – Display progress bar in stderr (default: false)
- ``--time`` – Measure and display execution time (default: false)

### Examples
1. Analyze the current repository with an extension filter:
```gitfame --extensions='.go,.md' --order-by=commits```
2. Output in JSON format for a specific commit:
```gitfame --revision=abc123 --format=json```
3. Analyze with a progress bar and language filter:
```gitfame --languages='go,markdown' --progress```

### Output Formats
#### Tabular
```
Name             Lines  Commits  Files  
GlebMoskalev     977    14       7  
AlexDeveloper    642    10       5  
DianaMaintainer  315    6        3  
```
#### CSV
```
Name,Lines,Commits,Files  
GlebMoskalev,977,14,7  
AlexDeveloper,642,10,5  
DianaMaintainer,315,6,3  
```
#### JSON
```json
[
  {"Name": "GlebMoskalev", "Lines": 977, "Commits": 14, "Files": 7},
  {"Name": "AlexDeveloper", "Lines": 642, "Commits": 10, "Files": 5},
  {"Name": "DianaMaintainer", "Lines": 315, "Commits": 6, "Files": 3}
]
```
#### JSON Lines
```json lines
{"Name": "GlebMoskalev", "Lines": 977, "Commits": 14, "Files": 7}
{"Name": "AlexDeveloper", "Lines": 642, "Commits": 10, "Files": 5}
{"Name": "DianaMaintainer", "Lines": 315, "Commits": 6, "Files": 3}
```

## Integration Tests
GitFame includes a comprehensive suite of integration tests to ensure the utility works as expected across various scenarios. These tests are located in the `test/integration` directory and are designed to validate the behavior of the `gitfame` binary with real Git repositories.
### Directory Structure

- **`test/integration/`**: Contains the integration test logic and supporting files.
    - **`gitfame_test.go`**: The main test file that defines the test suite and logic for running the `gitfame` binary against test cases.
    - **`testdata/`**: Directory holding test cases and sample repositories.
        - **`tests/`**: Subdirectory with individual test cases, each containing:
            - `description.yaml`: Defines the test name, command-line arguments, expected output format, and whether an error is expected.
            - `expected.out`: The expected output for the test case.
        - **`bundles/`**: Subdirectory with sample Git repositories used as test inputs.

### Running Tests
To execute the integration tests, run the following command from the project root:

```bash
go test ./test/integration -v
```