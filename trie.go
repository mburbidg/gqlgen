package main

type Trie[T comparable] struct {
	children map[T]*Trie[T]
	isWord   bool
}

func NewTrie[T comparable]() *Trie[T] {
	return &Trie[T]{
		children: make(map[T]*Trie[T]),
	}
}

func (t *Trie[T]) Insert(word []T) {
	if len(word) == 0 {
		t.isWord = true
		return
	}
	if n, ok := t.children[word[0]]; ok {
		n.Insert(word[1:])
	} else {
		t.children[word[0]] = NewTrie[T]()
		t.children[word[0]].Insert(word[1:])
	}
}
func (t *Trie[T]) Search(word []T) bool {
	if len(word) == 0 {
		return t.isWord
	}
	if n, ok := t.children[word[0]]; ok {
		return n.Search(word[1:])
	}
	return false
}

func (t *Trie[T]) Count() int {
	var cnt int
	for _, n := range t.children {
		cnt += n.Count()
	}
	if t.isWord {
		cnt += 1
	}
	return cnt
}
