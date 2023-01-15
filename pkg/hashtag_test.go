package pkg

import (
	ahocorasick "github.com/BobuSumisu/aho-corasick"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func buildComplexTrie() *ahocorasick.Trie {
	cleanerStrings := []string{
		"slon",
		"a",
		"scar",
		"et",
		"clean",
		"lean",
		"long",
		"scarp",
		"carpe",
		"cleaner",
		"leaner",
		"this",
		"s",
		"ane",
		"er",
		"i",
		"is",
		"carp",
		"scarpe",
		"an",
		"n",
		"le",
		"o",
		"cle",
		"th",
		"ar",
		"ean",
		"on",
	}
	return buildTrie(cleanerStrings)
}

func buildTrie(cleanerStrings []string) *ahocorasick.Trie {
	builder := ahocorasick.NewTrieBuilder()
	builder.AddStrings(cleanerStrings)

	// let's add every single letter that we want to treat as a word as well
	alphabet := []string{
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}
	builder.AddStrings(alphabet)
	return builder.Build()
}

func TestSingleWordMatches(t *testing.T) {
	trie := buildComplexTrie()
	s := "cleaner"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)

	allMatches := matches.AllMatches
	assert.Equal(t, "cleaner", allMatches[0][0].MatchString())
	assert.Equal(t, "clean", allMatches[0][1].MatchString())
	assert.Equal(t, "leaner", allMatches[1][0].MatchString())
	assert.Equal(t, "er", allMatches[5][0].MatchString())
	assert.Equal(t, "e", allMatches[5][1].MatchString())

	for _, match := range matches.AllMatches {
		for _, m := range match {
			t.Logf("match: pos: %d - %s", m.Pos(), m.MatchString())
		}
	}
}

func TestSingleLetterHashtag(t *testing.T) {
	trie := buildTrie([]string{})
	s := "a"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)
	hashtags := matches.ComputeHashTags(0)
	require.Equal(t, 1, len(hashtags))
	assert.Equal(t, "A", hashtags[0].String)
}

func TestTwoLetterHashtag(t *testing.T) {
	trie := buildTrie([]string{})
	s := "ab"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)
	var hashtags []*HashTag

	hashtags = matches.ComputeHashTags(1)
	require.Equal(t, 1, len(hashtags))
	assert.Equal(t, "B", hashtags[0].String)
	assert.Equal(t, 1, hashtags[0].Words)

	hashtags = matches.ComputeHashTags(0)
	require.Equal(t, 1, len(hashtags))
	assert.Equal(t, "AB", hashtags[0].String)
	assert.Equal(t, 2, hashtags[0].Words)
}

func TestTwoLetterSingleWordHashtag(t *testing.T) {
	trie := buildTrie([]string{"ab"})
	s := "ab"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)
	var hashtags []*HashTag

	hashtags = matches.ComputeHashTags(1)
	require.Equal(t, 1, len(hashtags))
	assert.Equal(t, "B", hashtags[0].String)
	assert.Equal(t, 1, hashtags[0].Words)

	hashtags = matches.ComputeHashTags(0)
	require.Equal(t, 2, len(hashtags))
	assert.Equal(t, "Ab", hashtags[0].String)
	assert.Equal(t, 1, hashtags[0].Words)
	assert.Equal(t, "AB", hashtags[1].String)
	assert.Equal(t, 2, hashtags[1].Words)
}

func TestTwoLetterTwoWordsHashtag(t *testing.T) {
	trie := buildTrie([]string{"abc", "ab", "bc"})
	s := "abc"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)
	var hashtags []*HashTag

	expected := []*HashTag{
		&HashTag{String: "Bc", Words: 1},
		&HashTag{String: "BC", Words: 2},
	}
	hashtags = matches.ComputeHashTags(1)

	require.Equal(t, 2, len(hashtags))
	for i, h := range hashtags {
		assert.Equal(t, expected[i].String, h.String)
		assert.Equal(t, expected[i].Words, h.Words)
	}

	hashtags = matches.ComputeHashTags(0)
	expected = []*HashTag{
		&HashTag{String: "Abc", Words: 1},
		&HashTag{String: "AbC", Words: 2},
		&HashTag{String: "ABc", Words: 2},
		&HashTag{String: "ABC", Words: 3},
	}
	require.Equal(t, len(expected), len(hashtags))
	for i, h := range hashtags {
		assert.Equal(t, expected[i].String, h.String)
		assert.Equal(t, expected[i].Words, h.Words)
	}
}

func TestTwoLetterTwoWordsHashtagIterative(t *testing.T) {
	trie := buildTrie([]string{"abc", "ab", "bc"})
	s := "abc"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)

	hashtags := matches.ComputeHashTagsIterative(0)
	expected := []*HashTag{
		&HashTag{String: "Abc", Words: 1},
		&HashTag{String: "AbC", Words: 2},
		&HashTag{String: "ABc", Words: 2},
		&HashTag{String: "ABC", Words: 3},
	}
	require.Equal(t, len(expected), len(hashtags))
	for i, h := range hashtags {
		assert.Equal(t, expected[i].String, h.String)
		assert.Equal(t, expected[i].Words, h.Words)
	}
}

func TestSingleWordHashtags(t *testing.T) {
	trie := buildTrie([]string{"cleaner", "clean", "leaner"})

	s := "cleaner"
	trieMatches := trie.MatchString(s)
	matches := NewStringMatches(s, trieMatches)

	hashtags := matches.ComputeHashTags(0)
	expected := []*HashTag{
		&HashTag{String: "Cleaner", Words: 1},
		&HashTag{String: "CLeaner", Words: 2},
		&HashTag{String: "CleanER", Words: 3},
		&HashTag{String: "CLEANER", Words: 7},
	}
	require.Equal(t, len(expected), len(hashtags))
	for i, h := range hashtags {
		assert.Equal(t, expected[i].String, h.String)
		assert.Equal(t, expected[i].Words, h.Words)
	}
}
