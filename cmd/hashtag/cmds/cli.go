package cmds

import (
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/wesen/majuscule/pkg"
	"time"
)

var ReplCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start a REPL",
	Run: func(cmd *cobra.Command, args []string) {
		dicts, err := cmd.Flags().GetStringSlice("dict")
		cobra.CheckErr(err)

		trie, err := pkg.BuildTrieFromFiles(dicts)
		cobra.CheckErr(err)

		// read strings from stdin
		// for each string, find all matches
		for {
			var s string
			_, err := fmt.Scanln(&s)
			if err != nil {
				break
			}

			start := time.Now()
			var hashTags []*pkg.HashTag
			var trieMatches []*ahocorasick.Match
			iterCount := 1
			for i := 0; i < iterCount; i++ {
				trieMatches = trie.MatchString(s)
			}
			elapsed := time.Since(start)
			log.Debug().Int64("duration_ns", elapsed.Nanoseconds()).
				Int("iterations", iterCount).
				Str("s", s).
				Int("trieMatches", len(trieMatches)).
				Msg("Aho-Corasick Match")

			matchedStrings := make(map[string]interface{})
			for _, m := range trieMatches {
				log.Trace().
					Int64("pos", m.Pos()).
					Str("match", m.MatchString()).
					Msg("Match")
				matchedStrings[string(m.Match())] = nil
			}

			//// code to print out for unit tests
			//for k := range matchedStrings {
			//	fmt.Printf("\"%s\",\n", k)
			//}

			start = time.Now()
			iterCount = 1
			for i := 0; i < iterCount; i++ {
				matches_ := pkg.ComputeMatches(s, trieMatches, nil)
				matches := pkg.NewStringMatches(s, matches_)
				hashTags = matches.SuggestHashtags()
			}
			elapsed = time.Since(start)
			log.Debug().Int64("duration_ns", elapsed.Nanoseconds()).
				Int("iterations", iterCount).
				Str("s", s).
				Int("hashTags", len(hashTags)).
				Msg("SuggestHashtags")

			// show at most 5 results
			for _, hashTag := range hashTags[:5] {
				fmt.Printf("%d - %s\n", hashTag.Words, hashTag.Tag)
			}
		}
	},
}

func init() {
}
