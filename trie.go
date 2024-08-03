package main

import "slices"

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

func (t *Trie[T]) Remove(word []T) {
	if len(word) == 0 && t.isWord {
		t.isWord = false
		return
	}
	if n, ok := t.children[word[0]]; ok {
		n.Remove(word[1:])
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

func (t *Trie[T]) VisitAllWords(visitor func(word []T)) {
	t.visit(t, []T{}, visitor)
}

func (t *Trie[T]) visit(tr *Trie[T], word []T, visitor func(word []T)) {
	if tr.isWord {
		visitor(word)
	}
	for k, v := range tr.children {
		childWord := slices.Concat(word, []T{k})
		t.visit(v, childWord, visitor)
	}
}
