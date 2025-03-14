package main

import (
	"fmt"
	"github.com/GlebMoskalev/gitfame/internal/git"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func init() {
	rootCmd.Flags().String("repository", ".", "")
	rootCmd.Flags().String("revision", "HEAD", "")
	rootCmd.Flags().String("exclude", "", "")
	rootCmd.Flags().String("restrict-to", "", "")
	rootCmd.Flags().String("languages", "", "")
	rootCmd.Flags().String("extensions", "", "")
	rootCmd.Flags().String("format", "tabular", "")
	rootCmd.Flags().String("order-by", "lines", "")
	rootCmd.Flags().String("use-committer", "false", "")
}

var rootCmd = &cobra.Command{
	Use: "main",
	Run: func(cmd *cobra.Command, args []string) {
		repository, err := cmd.Flags().GetString("repository")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		revision, err := cmd.Flags().GetString("revision")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		exclude, err := cmd.Flags().GetString("exclude")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		restrictTo, err := cmd.Flags().GetString("restrict-to")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		languages, err := cmd.Flags().GetString("languages")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		extensions, err := cmd.Flags().GetString("extensions")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		order, err := cmd.Flags().GetString("order-by")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		format, err := cmd.Flags().GetString("format")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		useCommitterStr, err := cmd.Flags().GetString("use-committer")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		useCommitter, err := strconv.ParseBool(useCommitterStr)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		git.CalculateStats(repository, revision, extensions, exclude, restrictTo, languages, order, format, useCommitter)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
