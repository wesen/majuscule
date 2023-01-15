package pkg

import (
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"github.com/rs/zerolog/log"
	"runtime"
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
