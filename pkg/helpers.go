package pkg

import (
	"bufio"
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"github.com/rs/zerolog/log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func FormatMemorySize(alloc uint64) string {
	// convert to GB
	allocGB := float64(alloc) / 1024 / 1024 / 1024
	return fmt.Sprintf("%.2f GB", allocGB)
}

func BuildTrieFromFiles(paths []string) (*ahocorasick.Trie, error) {
	builder := ahocorasick.NewTrieBuilder()
	log.Debug().Msgf("Loading dictionaries...")
	var err error
	for _, path := range paths {
		log.Debug().Str("file", path).Msg("Loading...")
		err = builder.LoadStrings(path)

		if err != nil {
			return nil, err
		}
	}

	log.Debug().Msg("Building trie...")
	trie := builder.Build()
	log.Debug().Msg("Built.")

	// print allocated memory size from garbage collector information
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Debug().Msgf("Allocated memory size: %s", FormatMemorySize(mem.Alloc))

	return trie, nil
}

func LoadWordFrequencies(path string) (map[string]int, error) {
	// Create an empty map to store the frequency of the words
	wordFrequency := make(map[string]int)

	// Open the file with the word frequencies
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Skip the first line (headers)
	scanner.Scan()

	// Loop through each line of the file
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		wordRegexp := regexp.MustCompile(`^[a-zA-Z]+$`)

		if len(fields) == 3 {
			// Check if fields[0] matches the regexp /^[a-zA-Z]+$/
			if !wordRegexp.MatchString(fields[0]) {
				continue
			}

			word := strings.ToLower(fields[0])
			frequency := fields[2]

			// convert frequency string to int (golang)
			freq, err := strconv.Atoi(frequency)
			if err != nil {
				return nil, err
			}

			// Add the word and frequency to the map
			if _, ok := wordFrequency[word]; !ok {
				wordFrequency[word] = freq
			} else {
				if wordFrequency[word] < freq {
					wordFrequency[word] = freq
				}
			}
		}
	}

	return wordFrequency, nil
}
