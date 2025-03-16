package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/GlebMoskalev/gitfame/internal/stats"
)

type cliOptions struct {
	repository   string
	revision     string
	exclude      string
	restrictTo   string
	languages    string
	extensions   string
	format       string
	orderBy      string
	useCommitter bool
	showProgress bool
}

func main() {
	Execute()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var options = &cliOptions{}

var rootCmd = &cobra.Command{
	Use:   "gitfame",
	Short: "Calculate git repository statistics",
	Run: func(cmd *cobra.Command, args []string) {
		stats.CalculateStats(
			options.repository,
			options.revision,
			options.extensions,
			options.exclude,
			options.restrictTo,
			options.languages,
			options.orderBy,
			options.format,
			options.useCommitter,
			options.showProgress,
		)
	},
}

func init() {
	rootCmd.Flags().StringVar(&options.repository, "repository", ".", "Path to git repository")
	rootCmd.Flags().StringVar(&options.revision, "revision", "HEAD", "Git revision to analyze")
	rootCmd.Flags().StringVar(&options.exclude, "exclude", "", "Exclude files matching pattern")
	rootCmd.Flags().StringVar(&options.restrictTo, "restrict-to", "", "Restrict analysis to files matching pattern")
	rootCmd.Flags().StringVar(&options.languages, "languages", "", "Filter by language")
	rootCmd.Flags().StringVar(&options.extensions, "extensions", "", "Filter by file extensions")
	rootCmd.Flags().StringVar(&options.format, "format", "tabular", "Output format (tabular, csv, json)")
	rootCmd.Flags().StringVar(&options.orderBy, "order-by", "lines", "Order results by (lines, commits, files)")
	rootCmd.Flags().BoolVar(&options.useCommitter, "use-committer", false, "Use committer instead of author")
	rootCmd.Flags().BoolVar(&options.showProgress, "progress", false, "Display progress bar during analysis")
}
