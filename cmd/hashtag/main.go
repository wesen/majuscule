package main

import (
	"github.com/spf13/cobra"
	"github.com/wesen/majuscule/cmd/hashtag/cmds"
)

var rootCmd = &cobra.Command{
	Use:   "hashtag",
	Short: "hashtag is a tool for finding hashtags in text",
}

func init() {
	rootCmd.AddCommand(cmds.ReplCmd)
	rootCmd.AddCommand(cmds.CompleteCmd)
	rootCmd.AddCommand(cmds.ServeCmd)

	wordLists := []string{
		"test_data/words",
		//"test_data/words.txt",
		//"test_data/google-10000-english-no-swears.txt",
	}
	rootCmd.PersistentFlags().StringSlice("dict", wordLists, "Dictionary file(s) to use")
	rootCmd.PersistentFlags().String("frequency", "test_data/1_2_all_freq.txt", "Frequency file to use")
}

func main() {
	_ = rootCmd.Execute()
}
