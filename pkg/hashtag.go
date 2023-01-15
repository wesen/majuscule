package pkg

import (
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"sort"
	"strings"
)

type StringMatches struct {
	String     string
	AllMatches [][]*ahocorasick.Match
	cache      [][]*HashTag
}

func NewStringMatches(s string, matches []*ahocorasick.Match) *StringMatches {
	matches_ := make([][]*ahocorasick.Match, len(s))

	for _, match := range matches {
		pos := match.Pos()

		if matches_[pos] == nil {
			matches_[pos] = make([]*ahocorasick.Match, 0)
		}
		matches_[pos] = append(matches_[pos], match)
	}

	// we sort the individual matches to have the longest one first (most salient)
	for _, ms_ := range matches_ {
		sort.Slice(ms_, func(i, j int) bool {
			return len(ms_[i].Match()) > len(ms_[j].Match())
		})
	}

	return &StringMatches{
		s,
		matches_,
		make([][]*HashTag, len(s)),
	}
}

type HashTag struct {
	String string
	Words  int
}

func (ht *HashTag) Score() int {
	return ht.Words
}

// ComputeHashTagsIterative is an iterative, non-recursive version of ComputeHashTags
// which will make it easier to provide bounded depth, hopefully
func (sm *StringMatches) ComputeHashTagsIterative(pos int) []*HashTag {
	ret := make([]*HashTag, 0)

	if pos >= len(sm.String) {
		return []*HashTag{
			&HashTag{
				"",
				0,
			},
		}
	}

	if sm.cache[pos] != nil {
		return sm.cache[pos]
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Words < ret[j].Words
	})

	sm.cache[pos] = ret

	return ret

}

// ComputeHashTags computes the best hashtags starting at a given position
func (sm *StringMatches) ComputeHashTags(pos int) []*HashTag {
	ret := make([]*HashTag, 0)

	if pos >= len(sm.String) {
		return []*HashTag{
			&HashTag{
				"",
				0,
			},
		}
	}

	if sm.cache[pos] != nil {
		return sm.cache[pos]
	}

	// we go through all the matches at the current position and try to build a hashtag
	// and then recurse into the suffix
	for _, match := range sm.AllMatches[pos] {
		s := match.MatchString()

		for _, suffix := range sm.ComputeHashTags(pos + len(s)) {
			// we try to capitalize the first letter of the suffix
			// and then add it to the current match

			// we should try to do something about single letter words here
			// we try not to capitalize single letter words
			var s_ string
			//if len(s) > 1 {
			//	if len(suffix.String) > 1 {
			//		s_ = capitalize(s) + capitalize(suffix.String)
			//	} else {
			//		s_ = capitalize(s) + suffix.String
			//	}
			//} else {
			s_ = capitalize(s) + suffix.String
			//}

			ret = append(ret, &HashTag{
				String: s_,
				Words:  1 + suffix.Words,
			})
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Words < ret[j].Words
	})

	sm.cache[pos] = ret

	return ret
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// SuggestHashtags using a DP approach to computing possible hashtags
// It keeps track of the best result starting at a certain position.
// A best hashtag is the one that uses the least capitalizations to cover a given area.
func (sm *StringMatches) SuggestHashtags() []*HashTag {
	hashTags := sm.ComputeHashTags(0)

	// sort hashTags by Words
	sort.Slice(hashTags, func(i, j int) bool {
		return hashTags[i].Words < hashTags[j].Words
	})

	return hashTags
}
