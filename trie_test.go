package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTrie_InsertSearch(t *testing.T) {
	trie := NewTrie[int]()
	trie.Insert([]int{1, 2, 3})
	assert.True(t, trie.Search([]int{1, 2, 3}))
	assert.False(t, trie.Search([]int{1, 2}))
	assert.False(t, trie.Search([]int{1, 2, 3, 4}))
	assert.False(t, trie.Search([]int{1}))
	trie.Insert([]int{1})
	assert.True(t, trie.Search([]int{1}))
	assert.False(t, trie.Search([]int{1, 2}))
	assert.False(t, trie.Search([]int{1, 2, 3, 4}))
	trie.Insert([]int{1, 2, 3, 4})
	assert.True(t, trie.Search([]int{1, 2, 3, 4}))
	assert.False(t, trie.Search([]int{1, 2}))
	trie.Insert([]int{2, 3, 4})
	assert.False(t, trie.Search([]int{1, 2}))
	assert.True(t, trie.Search([]int{2, 3, 4}))
	assert.False(t, trie.Search([]int{2, 3, 4, 5}))
	assert.False(t, trie.Search([]int{2, 3}))
}

func TestTrie_Count(t *testing.T) {
	trie := NewTrie[int]()
	trie.Insert([]int{1, 2, 3})
	assert.Equal(t, 1, trie.Count())
	trie.Insert([]int{1, 2, 3, 4})
	assert.Equal(t, 2, trie.Count())
	trie.Insert([]int{1})
	assert.Equal(t, 3, trie.Count())
	trie.Insert([]int{1, 2, 3})
	assert.Equal(t, 3, trie.Count())
}
