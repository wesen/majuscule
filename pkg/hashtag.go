package pkg

import (
	"fmt"
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"math"
	"sort"
	"strings"
)

type Match struct {
	Match string
	Pos   int
	Score float64
}

func (m *Match) String() string {
	return fmt.Sprintf("%s (%d) - %f", m.Match, m.Pos, m.Score)
}

type StringMatches struct {
	String     string
	AllMatches [][]*Match
	cache      [][]*HashTag
}

func WordScore(word string, frequency map[string]int) float64 {
	// frequency is frequency / million

	l := float64(len(word))

	freq := 0
	if frequency != nil {
		f, ok := frequency[word]
		if ok {
			freq = f
		}
	}

	lengthWeight := 1.0
	frequencyWeight := 800.0

	freqFactor := (float64(freq) / 1000000.0) * frequencyWeight
	lengthFactor := l * l * lengthWeight
	score := lengthFactor + freqFactor
	return score
}

func NewStringMatches(s string, matches [][]*Match) *StringMatches {

	// we sort the individual matches to have the longest one first (most salient)

	return &StringMatches{
		s,
		matches,
		make([][]*HashTag, len(s)),
	}
}

func ComputeMatches(s string, matches []*ahocorasick.Match, frequency map[string]int) [][]*Match {
	matches_ := make([][]*Match, len(s))

	for _, match := range matches {
		pos := match.Pos()

		if matches_[pos] == nil {
			matches_[pos] = make([]*Match, 0)
		}
		matches_[pos] = append(matches_[pos], &Match{
			Match: match.MatchString(),
			Pos:   int(match.Pos()),
			Score: WordScore(match.MatchString(), frequency),
		})
	}

	for _, ms_ := range matches_ {
		sort.Slice(ms_, func(i, j int) bool {
			return ms_[i].Score > ms_[j].Score
		})
	}
	return matches_
}

type HashTag struct {
	Words  []string
	Scores []float64
}

func (ht *HashTag) Tag() string {
	return strings.Join(ht.Words, "")
}

func NewHashTag(words []string, scores []float64) *HashTag {
	return &HashTag{
		Words:  words,
		Scores: scores,
	}
}

// Score computes the score for the hashtag, lower is better
func (ht *HashTag) Score() float64 {
	// the sum of the scores
	score := 0.0
	for _, s := range ht.Scores {
		score += s
	}
	return score / float64(len(ht.Scores))
}

func (ht *HashTag) String() string {
	scoresString := make([]string, len(ht.Scores))
	for i, score := range ht.Scores {
		scoresString[i] = fmt.Sprintf("%f", score)
	}
	return fmt.Sprintf("%s (%d) [%s]", ht.Tag, ht.Words, strings.Join(scoresString, ","))
}

func (ht *HashTag) AppendMatch(match string, score float64) *HashTag {
	return NewHashTag(
		append(ht.Words, capitalize(match)),
		append(ht.Scores, score),
	)
}

func (ht *HashTag) Prepend(match string, score float64) *HashTag {
	return NewHashTag(
		append([]string{capitalize(match)}, ht.Words...),
		append([]float64{score}, ht.Scores...),
	)
}
func (ht *HashTag) AppendMatchWithSuffix(match string, matchScore float64, suffix *HashTag) *HashTag {
	return NewHashTag(
		append(append(ht.Words, capitalize(match)), suffix.Words...),
		append(append(ht.Scores, matchScore), suffix.Scores...),
	)
}

type toGoStackEntry struct {
	prefix      *HashTag
	matchString string
	pos         int
	score       float64
}

func NewToGoStackEntry(prefix *HashTag, match *Match) *toGoStackEntry {
	return &toGoStackEntry{
		prefix,
		match.Match,
		match.Pos,
		match.Score,
	}
}

func (e *toGoStackEntry) String() string {
	return fmt.Sprintf("{prefix: %s, match %s (%d)}", e.prefix, e.matchString, e.pos)
}

type toGoStack []*toGoStackEntry

func (s *toGoStack) Len() int {
	return len(*s)
}

func (s *toGoStack) Push(e *toGoStackEntry) {
	*s = append(*s, e)
}

func (s *toGoStack) Pop() *toGoStackEntry {
	if len(*s) == 0 {
		return nil
	}

	ret := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return ret
}

func (sm *StringMatches) createInitialToGoStack() *toGoStack {
	stack := make(toGoStack, 0)

	// we need to append sm.AllMatches[0] in reverse order to the list,
	// so that we have the highest weight on top of the stack
	if len(sm.AllMatches) > 0 {
		for i := len(sm.AllMatches[0]) - 1; i >= 0; i-- {
			match := sm.AllMatches[0][i]
			stack.Push(NewToGoStackEntry(NewHashTag([]string{}, []float64{}), match))
		}
	}

	return &stack
}

type hashTagStep struct {
	cache         map[int][]*HashTag
	cacheMaxScore map[int]float64
	toGoStack     *toGoStack
	curEntry      *toGoStackEntry
	curPos        int
	matchString   string
	nextPos       int
}

// ComputeHashTagsIterative is an iterative, non-recursive version of ComputeHashTags
// which will make it easier to provide bounded depth, hopefully.
// `maxResults` is the maximum number of results to return.
//
// while we are at it, let's introduce a `maxResults` of results to return.
// we can start cutting of the depth if we have more than maxCount entries in
// `ret`, because anything that will have a higher score won't make it in there anyway
//
// A maxResults of 0 means no limit
func (sm *StringMatches) ComputeHashTagsIterative(maxResults int) []*HashTag {
	// entries in ret should be sorted by Score()
	ret := make([]*HashTag, 0)

	// we have two stacks, one for how far we are in the string,
	// and one for the possibilities that we are exploring

	// this is what we have to go on
	// for each step that we have to do down, we need not just the match to process,
	// but the history of how we got there. We need a new struct
	toGo := sm.createInitialToGoStack()

	cache := make(map[int][]*HashTag, len(sm.String))

	// we also store how many words max we have for each cache entry, in case we need to iterate
	// deeper.
	//
	// BRAINSTORM: we also need to store the fact that we might have explored this entry entirely,
	// instead of having cutoff our search because of a threshold depth.
	cacheMaxScore := make(map[int]float64, len(sm.String))

	// the maximum number of words new entries should have,
	// this is the maximum number of words in ret, if ret is bigger than maxResults,
	// other it is just the length of the string, since each letter can be a single word
	// TODO: this should actually be the score, but for, score is just word count
	// set maxScore to infinity
	maxScore := math.Inf(1)

	recordedSteps := make([]*hashTagStep, 0)

	appendResult := func(cur *toGoStackEntry, suffix *HashTag) {
		matchString := cur.matchString
		curPos := cur.pos

		newHashTag := cur.prefix.AppendMatchWithSuffix(matchString, cur.score, suffix)
		suffixHashTag := suffix.Prepend(capitalize(matchString), cur.score)

		_, ok := cache[curPos]
		if !ok {
			cache[curPos] = make([]*HashTag, 0)
		}
		// BRAINSTORM: we append this to our cached entry,
		// but really the cache should only be updated once
		// we "step back" a level. This might be fine however, since we only
		// ever go to further suffixes from here
		cache[curPos] = append(cache[curPos], suffixHashTag)

		// now update the max words count for this position
		if suffixHashTag.Score() > cacheMaxScore[curPos] {
			cacheMaxScore[curPos] = suffixHashTag.Score()
		}

		// insert the new hashtag into ret, at the right position by Score()
		ret = insertSortedByScore(ret, newHashTag)

		if maxResults > 0 && len(ret) > maxResults {
			// if we have more than maxResults score, then there is no need to compute matches
			// that have a higher score than the entry at maxResults
			maxScore = ret[maxResults-1].Score()
		}
	}

	maxLen := 400

	for {
		if toGo.Len() == 0 {
			break
		}

		if len(ret) > maxLen {
			break
		}

		// pop off the first of the toGo matches
		cur := toGo.Pop()

		// BRAINSTORM: what does it mean to explore a pos completely?
		// it means that the next entry in the toGo is higher, and we haven't
		// cut off our search. So we need to recognize when we are going back up a level
		//

		matchString := cur.matchString
		curPos := cur.pos
		nextPos := curPos + len(matchString)

		// record step
		//recordedStep := recordStep(cache, toGo, cacheMaxScore, cur, curPos, matchString, nextPos)
		//recordedSteps = append(recordedSteps, recordedStep)

		// we now explore all the suffix matches at the newPos

		// if we are at the end of the string, we can now add a new
		// hashtag to the cache at cur.pos, and add to the results
		// (sorted by score)
		if nextPos >= len(sm.String) {
			appendResult(cur, NewHashTag([]string{}, []float64{}))
		} else {
			// we now "recurse" by adding all the matches at the next position to the toGo,
			// if they could potentially lead to lower scores than maxScore

			// BRAINSTORM: can we reused cached results if their score is lower than maxScore?
			//
			// we can definitely have cached entries that have been previously cutoff because
			// we could have gotten there with a long first match but then many short matches, for example
			// cleaner12345tombstone where cleaner would be tackled first, but then gotten 5 single letter words
			// before matching is and great, and potentially only matching tombstone but not tomb, stone.
			// however, we could, if for some reason "er12345" is a word, get there later with 2 matches
			// clean, er12345 and thus have enough "headroom" to match tomb, stone.
			//
			// I don't think we need to really special case this, because in the case of shorter matches, it
			// means that the next potential matches are either already cached with a proper maxScore,
			// or we'd need to recurse into those anyway too.
			//
			// this check for maxScore should be done at the beginning of the loop, no need to be clever

			_ = maxScore
			//if len(cache[nextPos]) > 0 && cacheMaxScore[nextPos] < maxScore {
			//	// BRAINSTORM: maybe in the first pass I shouldn't worry about the maxScore trick,
			//	// and get the iterative version correct.
			//
			//}

			if len(cache[nextPos]) > 0 {
				for _, suffixHashTag := range cache[nextPos] {
					appendResult(cur, suffixHashTag)
				}
			} else {
				// this means we need to recurse in for real now, by adding the matches
				// to toGo.
				// append sm.AllMatches[nextPos] in reverse order to have the highest weight on top
				for i := len(sm.AllMatches[nextPos]) - 1; i >= 0; i-- {
					match := sm.AllMatches[nextPos][i]
					newHashTag := cur.prefix.AppendMatch(capitalize(matchString), cur.score)
					nextToGo := NewToGoStackEntry(newHashTag, match)
					toGo.Push(nextToGo)
				}
			}
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		// score here could be score of the individual words, but inverse to the total count of words, or something like that ?
		// or maybe just average score of all words / number of words ?
		if ret[i].Score() == ret[j].Score() {
			return ret[i].Tag() > ret[j].Tag()
		} else {
			return ret[i].Score() > ret[j].Score()
		}
	})

	_ = recordedSteps

	return ret

}

func recordStep(
	cache map[int][]*HashTag,
	toGo *toGoStack,
	cacheMaxScore map[int]float64,
	cur *toGoStackEntry,
	curPos int,
	matchString string,
	nextPos int,
) *hashTagStep {
	cacheCopy := make(map[int][]*HashTag, len(cache))
	for k, v := range cache {
		vCopy := make([]*HashTag, len(v))
		copy(vCopy, v)
		cacheCopy[k] = vCopy
	}
	toGoStackCopy := make(toGoStack, toGo.Len())
	copy(toGoStackCopy, *toGo)

	cacheMaxScoreCopy := make(map[int]float64, len(cacheMaxScore))
	for k, v := range cacheMaxScore {
		cacheMaxScoreCopy[k] = v
	}

	recordedStep := &hashTagStep{
		cache:         cacheCopy,
		cacheMaxScore: cacheMaxScoreCopy,
		toGoStack:     &toGoStackCopy,
		curEntry:      cur,
		curPos:        curPos,
		matchString:   matchString,
		nextPos:       nextPos,
	}
	return recordedStep
}

func insertSortedByScore(ret []*HashTag, tag *HashTag) []*HashTag {
	if len(ret) == 0 {
		ret = append(ret, tag)
		return ret
	}

	// we need to find the right position to insert the tag
	// we can do a binary search, since the slice is sorted by Score()
	insertPos := sort.Search(len(ret), func(i int) bool {
		return ret[i].Score() >= tag.Score()
	})

	ret = append(ret, nil)
	copy(ret[insertPos+1:], ret[insertPos:])
	ret[insertPos] = tag

	return ret
}

// ComputeHashTags computes the best hashtags starting at a given position
func (sm *StringMatches) ComputeHashTags(pos int) []*HashTag {
	ret := make([]*HashTag, 0)

	if pos >= len(sm.String) {
		return []*HashTag{
			NewHashTag([]string{}, []float64{}),
		}
	}

	if sm.cache[pos] != nil {
		return sm.cache[pos]
	}

	// we go through all the matches at the current position and try to build a hashtag
	// and then recurse into the suffix
	for _, match := range sm.AllMatches[pos] {
		s := match.Match

		for _, suffix := range sm.ComputeHashTags(pos + len(s)) {
			// we try to capitalize the first letter of the suffix
			// and then add it to the current match
			ret = append(ret, suffix.Prepend(capitalize(s), match.Score))
		}
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Score() > ret[j].Score()
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
	hashTags := sm.ComputeHashTagsIterative(0)

	return hashTags
}
