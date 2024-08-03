package main

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"slices"
)

type kpathGenerator struct {
	grammar          *node
	rules            map[string]*node
	minCharStringLen int
}

func newKpathGenerator(grammar *node) *kpathGenerator {
	transformSeeTheRules(grammar)
	transformRepeat(grammar)
	transformAlt(grammar)
	assignId(grammar)
	nameRhs(grammar)

	g := &kpathGenerator{grammar: grammar}
	g.buildRuleMap()

	return g
}

func (g *kpathGenerator) generate(k int) {
	tr := g.buildKpaths(k)
	g.printKpaths(tr)
}

func (g *kpathGenerator) printKpaths(tr *Trie[int]) {
	fmt.Printf("total kpaths: %d\n", tr.Count())
	tr.VisitAllWords(func(word []int) {
		fmt.Println(word)
	})
}

func (g *kpathGenerator) buildRuleMap() {
	g.rules = make(map[string]*node)
	walk(g.grammar, func(n *node) {
		if n.kind == bnfDefKind {
			g.rules[n.name] = n.children[0] // rhs
		}
	})
	g.rules["character representation"] = &node{
		kind:       fnKind,
		generateFn: g.generateCharacterRepresentation,
	}
	g.rules["string literal character"] = &node{
		kind:       fnKind,
		generateFn: g.generateStringLiteralCharacter,
	}
	g.rules["identifier start"] = &node{
		kind:       fnKind,
		generateFn: g.generateIdentifierStart,
	}
	g.rules["identifier extend"] = &node{
		kind:       fnKind,
		generateFn: g.generateIdentifierExtend,
	}
	g.rules["whitespace"] = &node{
		kind:       fnKind,
		generateFn: g.generateWhitespace,
	}
	g.rules["truncating whitespace"] = &node{
		kind:       fnKind,
		generateFn: g.generateTruncatingWhitespace,
	}
	g.rules["bidirectional control character"] = &node{
		kind:       fnKind,
		generateFn: g.generateBidirectionControlCharacter,
	}
	g.rules["simple comment character"] = &node{
		kind:       fnKind,
		generateFn: g.generateSimpleCommentCharacter,
	}
	g.rules["bracketed comment contents"] = &node{
		kind:       fnKind,
		generateFn: g.generateBracketedCommentContents,
	}
	g.rules["newline"] = &node{
		kind:       fnKind,
		generateFn: g.generateNewline,
	}
	g.rules["other digit"] = &node{
		kind:       fnKind,
		generateFn: g.generateOtherDigit,
	}
	g.rules["other language character"] = &node{
		kind:       fnKind,
		generateFn: g.generateOtherLanguageCharacter,
	}
	g.rules["single quoted character sequence"] = &node{
		kind:       fnKind,
		generateFn: g.generateSingleQuotedCharacterSequence,
	}
	g.rules["double quoted character sequence"] = &node{
		kind:       fnKind,
		generateFn: g.generateDoubleQuotedCharacterSequence,
	}
	g.rules["accent quoted character sequence"] = &node{
		kind:       fnKind,
		generateFn: g.generateAccentQuotedCharacterSequence,
	}
}

func (g *kpathGenerator) buildKpaths(k int) *Trie[int] {
	tr := NewTrie[int]()
	walk(g.grammar, func(n *node) {
		if n.kind != grammarKind {
			g.addKpaths(n, k, []int{}, tr)
		}
	})
	return tr
}

func (g *kpathGenerator) addKpaths(n *node, k int, kpath []int, tr *Trie[int]) {
	if len(kpath) > k {
		tr.Insert(kpath)
		return
	}
	childKPath := slices.Concat(kpath, []int{n.id})
	switch n.kind {
	case kwKind, terminalSymbolKind, fnKind:
		return
	case bnfKind:
		child := g.rules[n.name]
		g.addKpaths(child, k, childKPath, tr)
	case rhsKind, groupKind, optKind, repeatKind, altKind, bnfDefKind:
		for _, n := range n.children {
			g.addKpaths(n, k, childKPath, tr)
		}
	}
}

func (g *kpathGenerator) generateSingleQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(g.minCharStringLen, 5)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"\"`", []string{"''"})
		if err != nil {
			panic(err)
		}
		return "@\"" + guts + "\""
	} else {
		guts, err := g.randString(cnt, charset+"\"`", append(escapeSequences, "''"))
		if err != nil {
			panic(err)
		}
		return "'" + guts + "'"
	}
}

func (g *kpathGenerator) generateDoubleQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(g.minCharStringLen, 5)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"'`", []string{`""`})
		if err != nil {
			panic(err)
		}
		return "@\"" + guts + "\""
	} else {
		guts, err := g.randString(cnt, charset+"'`", append(escapeSequences, `""`))
		if err != nil {
			panic(err)
		}
		return "\"" + guts + "\""
	}
}

func (g *kpathGenerator) generateAccentQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(g.minCharStringLen, 5)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"\"'", []string{"``"})
		if err != nil {
			panic(err)
		}
		return "@`" + guts + "`"
	} else {
		guts, err := g.randString(cnt, charset+"\"'", append(escapeSequences, "``"))
		if err != nil {
			panic(err)
		}
		return "`" + guts + "`"
	}
}

func (g *kpathGenerator) generateCharacterRepresentation(n *node) string {
	r, err := randChar(charset)
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *kpathGenerator) generateStringLiteralCharacter(n *node) string {
	return "somerandomstring"
}

func (g *kpathGenerator) generateIdentifierStart(n *node) string {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ")
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *kpathGenerator) generateIdentifierExtend(n *node) string {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ0123456789")
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *kpathGenerator) generateWhitespace(n *node) string {
	return " "
}

func (g *kpathGenerator) generateTruncatingWhitespace(n *node) string {
	return " "
}

func (g *kpathGenerator) generateBidirectionControlCharacter(n *node) string {
	return ""
}

func (g *kpathGenerator) generateSimpleCommentCharacter(n *node) string {
	return ""
}

func (g *kpathGenerator) generateBracketedCommentContents(n *node) string {
	return ""
}

func (g *kpathGenerator) generateNewline(n *node) string {
	return "\n"
}

func (g *kpathGenerator) generateOtherDigit(n *node) string {
	return ""
}

func (g *kpathGenerator) generateOtherLanguageCharacter(n *node) string {
	return ""
}

func (g *kpathGenerator) randomRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func (g *kpathGenerator) randString(n int, charset string, escapeSequences []string) (string, error) {
	s, r := make([]rune, n), []rune(charset)
	for i := range s {
		p, err := crand.Prime(crand.Reader, len(r))
		if err != nil {
			return "", fmt.Errorf("random string n %d: %w", n, err)
		}
		x, y := p.Uint64(), uint64(len(r))
		// fmt.Printf("x: %d y: %d\tx %% y = %d\trandom[%d] = %q\n", x, y, x%y, x%y, string(r[x%y]))
		s[i] = r[x%y]
	}
	str := string(s)
	if len(escapeSequences) > 0 && g.randomRange(0, 100) > 80 {
		// include an escaped character only 20% of the time.
		m := g.randomRange(0, 100)
		if m >= 80 {
			str = str + escapeSequences[g.randomRange(0, len(escapeSequences))]
		}
	}
	return str, nil
}

func (g *kpathGenerator) enterLeave(n *node) func() {
	n.cnt++
	if n.kind == rhsKind && n.name == "delimited identifier" {
		g.minCharStringLen = 1
	}
	return func() {
		n.cnt--
		if n.kind == rhsKind && n.name == "delimited identifier" {
			g.minCharStringLen = 0
		}
	}
}
