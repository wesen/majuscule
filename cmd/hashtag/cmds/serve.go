package cmds

import (
	"embed"
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/wesen/majuscule/pkg"
	"io/fs"
	"net/http"
	"time"
)

type Server struct {
	trie      *ahocorasick.Trie
	port      string
	frequency map[string]int
}

type AhoCorasickMatch struct {
	Pos   int     `json:"pos"`
	Word  string  `json:"word"`
	Score float64 `json:"score"`
}

type HashTag struct {
	Tag    string    `json:"tag"`
	Score  float64   `json:"score"`
	Words  []string  `json:"words"`
	Scores []float64 `json:"scores"`
}

type CompleteResponse struct {
	Input              string              `json:"input"`
	Count              int                 `json:"count"`
	Hashtags           []*HashTag          `json:"hashtags"`
	Matches            []*AhoCorasickMatch `json:"matches,omitempty"`
	MatchDuration_ns   int64               `json:"match_duration_ns"`
	SuggestDuration_ns int64               `json:"suggest_duration_ns"`
}

type CompleteResponses []CompleteResponse

type CompleteRequest struct {
	Inputs []string `json:"inputs"`
	Count  int      `json:"count"`
	Debug  bool     `json:"debug"`
}

func (s *Server) computeHashtags(input string, count int) CompleteResponse {
	results := CompleteResponse{
		Input:    input,
		Count:    count,
		Hashtags: make([]*HashTag, 0),
		Matches:  make([]*AhoCorasickMatch, 0),
	}

	// cheap ass limiting
	if len(input) > 60 {
		return results
	}

	start := time.Now()

	trieMatches := s.trie.MatchString(input)
	matches_ := pkg.ComputeMatches(input, trieMatches, s.frequency)
	elapsed := time.Since(start)

	log.Debug().Int64("duration_ns", elapsed.Nanoseconds()).
		Str("input", input).
		Int("trieMatches", len(trieMatches)).
		Msg("Match")

	for _, m := range matches_ {
		for _, w := range m {
			results.Matches = append(results.Matches, &AhoCorasickMatch{
				Pos:   w.Pos,
				Word:  w.Match,
				Score: w.Score,
			})
		}
	}

	//// code to print out for unit tests
	//for k := range matchedStrings {
	//	fmt.Printf("\"%s\",\n", k)
	//}

	start = time.Now()
	matches := pkg.NewStringMatches(input, matches_)
	hashTags := matches.SuggestHashtags()

	for i, h := range hashTags {
		if i > count {
			break
		}
		results.Hashtags = append(results.Hashtags, &HashTag{
			Tag:    h.Tag(),
			Score:  h.Score(),
			Words:  h.Words,
			Scores: h.Scores,
		})
	}

	elapsed = time.Since(start)
	results.SuggestDuration_ns = elapsed.Nanoseconds()

	return results
}

//go:embed web/*
var webFS embed.FS

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		return false
	}

	// check if indexing is allowed
	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

func EmbedFolder(fsEmbed embed.FS, targetPath string, index bool) static.ServeFileSystem {
	subFS, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(subFS),
		indexes:    index,
	}
}

func (s *Server) Run() error {
	router := gin.Default()

	router.GET("/complete", func(c *gin.Context) {
		countString := c.DefaultQuery("count", "5")
		count := 5
		_, err := fmt.Sscanf(countString, "%d", &count)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count"})
			return
		}

		debug := c.DefaultQuery("debug", "false")

		input := c.Query("input")
		response := s.computeHashtags(input, count)

		if debug != "true" {
			response.Matches = nil
		}
		c.JSON(http.StatusOK, response)
	})

	router.POST("/complete", func(c *gin.Context) {
		var req CompleteRequest
		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		responses := make([]CompleteResponse, len(req.Inputs))
		for i, input := range req.Inputs {
			responses[i] = s.computeHashtags(input, req.Count)
		}

		if !req.Debug {
			for i := range responses {
				responses[i].Matches = nil
			}
		}

		c.JSON(http.StatusOK, responses)
	})

	fs := EmbedFolder(webFS, "web", true)
	router.Use(static.Serve("/", fs))

	addr := ":" + s.port
	log.Info().Str("port", s.port).Msg("Starting server")
	return router.Run(addr)
}

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the hashtag server",
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetString("port")

		dicts, err := cmd.Flags().GetStringSlice("dict")
		cobra.CheckErr(err)

		frequencyPath, err := cmd.Flags().GetString("frequency")
		cobra.CheckErr(err)

		trie, err := pkg.BuildTrieFromFiles(dicts)
		cobra.CheckErr(err)

		frequency, err := pkg.LoadWordFrequencies(frequencyPath)
		cobra.CheckErr(err)

		s := &Server{
			trie:      trie,
			frequency: frequency,
			port:      port,
		}

		err = s.Run()
		cobra.CheckErr(err)
	},
}

var GrpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "Starts the hashtag server",
	Run: func(cmd *cobra.Command, args []string) {

		//s := grpc.NewServer()
		//grpc2.RegisterCompleteServer(s, &server{})
	},
}

func init() {
	ServeCmd.Flags().StringP("port", "p", "8080", "Port to listen on")
}
