package main

type kpathGenerator struct {
	grammar *node
}

func newKpathGenerator(grammar *node) *kpathGenerator {
	transformSeeTheRules(grammar)
	transformRepeat(grammar)
	transformAlt(grammar)
	assignId(grammar)
	nameRhs(grammar)

	return &kpathGenerator{grammar: grammar}
}

func (g *kpathGenerator) generate() {
}

func (g *kpathGenerator) buildKpaths(k int) {
	tr := NewTrie[int]()
	walk(g.grammar, func(n *node) {
		g.addKpaths(n, k, tr)
	})
}

func (g *kpathGenerator) addKpaths(n *node, k int, tr *Trie[int]) {
}
