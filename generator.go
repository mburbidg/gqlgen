package main

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
)

const (
	singleQuotedCharacterSequence = iota
	doubleQuotedCharacterSequence
	accentQuotedCharacterSequence
)

type generator struct {
	rules                       map[string]*node
	grammar                     *node
	maxRevist                   int
	includeDelimitedIdentifiers bool
}

var errRecursionLevelExceeded = errors.New("recursion level exceeded")

func newGenerator(grammar *node, maxRevisit int) *generator {
	g := &generator{grammar: grammar, maxRevist: maxRevisit}

	g.transformSeeTheRules(g.grammar)
	g.transformRepeat(g.grammar)
	g.transformAlt(g.grammar)
	g.condenseRhs(g.grammar)
	g.assignId(g.grammar)
	g.buildRuleMap()

	return g
}

func (g *generator) walk(n *node, visitor func(n *node)) {
	visitor(n)
	for _, child := range n.children {
		g.walk(child, visitor)
	}
}

func (g *generator) generate(startRule string, tree *node, verbose bool) string {
	if start, ok := g.rules[startRule]; ok {
		for {
			s, err := g.generateNode(start)
			if err == nil {
				return s
			} else {
				if verbose {
					fmt.Printf("%s\n", err)
				}
			}
		}
	} else {
		panic(fmt.Sprintf("unknown start rule: %s", startRule))
	}
}

func (g *generator) buildRuleMap() {
	g.rules = make(map[string]*node)
	g.walk(g.grammar, func(n *node) {
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

func (g *generator) transformRepeat(n *node) {
	var prev *node
	toRemove := []*node{}
	for _, child := range n.children {
		if child.kind == repeatKind {
			child.children = append(child.children, prev)
			prev.parent = child
			toRemove = append(toRemove, prev)
		} else {
			g.transformRepeat(child)
		}
		prev = child
	}
	for _, child := range toRemove {
		g.removeChild(n, child)
	}
}

func (g *generator) removeChild(n, childToRemove *node) {
	c := []*node{}
	for _, child := range n.children {
		if child.kind != childToRemove.kind {
			c = append(c, child)
		}
	}
	n.children = c
}

func (g *generator) assignId(root *node) {
	var nextId int
	g.walk(root, func(n *node) {
		n.id = nextId
		nextId++
	})
}

func (g *generator) resetCnts(root *node) {
	g.walk(root, func(n *node) {
		if n.cnt != 0 {
			fmt.Printf("cnt for (%s, %d) not zero: %d\n", n.kind, n.id, n.cnt)
		}
		n.cnt = 0
	})
}

func (g *generator) transformAlt(n *node) {
	for _, child := range n.children {
		g.transformAlt(child)
	}
	if len(n.children) > 1 && n.children[0].kind == altKind {
		alt := &node{kind: altKind, parent: n}
		for _, child := range n.children {
			if child.kind == altKind {
				if len(child.children) > 1 {
					alt.children = append(alt.children, child)
					child.kind = groupKind
				} else {
					child.children[0].parent = alt
					alt.children = append(alt.children, child.children[0])
				}
			} else {
				panic("alt mixed with other nodes")
			}
		}
		n.children = []*node{alt}
	}
}

func (g *generator) transformSeeTheRules(n *node) {
	toRemove := []*node{}
	for _, child := range n.children {
		if child.kind == seeTheRulesKind {
			toRemove = append(toRemove, child)
		} else {
			g.transformSeeTheRules(child)
		}
	}
	for _, child := range toRemove {
		g.removeChild(n, child)
	}
}

func (g *generator) printNode(n *node, indent string) {
	if n.kind == kwKind {
		fmt.Printf("%s%s(%d, %s)\n", indent, n.kind, n.id, n.value)
	} else if n.kind == bnfKind || n.kind == bnfDefKind {
		fmt.Printf("%s%s(%d, %s)\n", indent, n.kind, n.id, n.name)
	} else {
		fmt.Printf("%s%s(%d)\n", indent, n.kind, n.id)
	}
	for _, child := range n.children {
		g.printNode(child, indent+"  ")
	}
}

func (g *generator) condenseRhs(n *node) {
	for _, child := range n.children {
		if child.kind == bnfDefKind {
			rhs := child.children[0]
			if len(rhs.children) == 1 {
				child.children[0] = rhs.children[0]
			} else {
				rhs.kind = groupKind
			}
		} else {
			panic(fmt.Sprintf("expecting %s\n", bnfDefKind))
		}
	}
}

func (g *generator) generateNode(n *node) (string, error) {
	defer g.enterLeave(n)()
	if n.cnt > g.maxRevist {
		return "", errRecursionLevelExceeded
	}
	//fmt.Printf("generate node: kind=%s, id=%d\n", n.kind, n.id)
	switch n.kind {
	case altKind:
		return g.generateAlt(n)
	case bnfKind:
		return g.generateBnf(n)
	case optKind:
		return g.generateOpt(n)
	case groupKind:
		return g.generateGroup(n)
	case repeatKind:
		return g.generateRepeat(n)
	case terminalSymbolKind:
		return g.generateTerminalSymbol(n)
	case kwKind:
		return g.generateKw(n)
	case fnKind:
		return n.generateFn(n), nil
	}
	return "", nil
}

func (g *generator) generateBnf(n *node) (string, error) {
	if n, ok := g.rules[n.name]; ok {
		return g.generateNode(n)
	} else {
		panic("rule not found: " + n.name)
	}
	return "", nil
}

func (g *generator) generateAlt(n *node) (string, error) {
	i := g.randomRange(0, len(n.children))
	//fmt.Printf("alt id=%d, i=%d, child=%d\n", n.id, i, n.children[0].id)
	return g.generateNode(n.children[i])
}

func (g *generator) generateOpt(n *node) (string, error) {
	result := ""
	i := g.randomRange(0, 2)
	if i == 1 {
		for _, child := range n.children {
			s, err := g.generateNode(child)
			if err != nil {
				return "", err
			}
			result += s
		}
	}
	return result, nil
}

func (g *generator) generateGroup(n *node) (string, error) {
	result := ""
	for _, child := range n.children {
		s, err := g.generateNode(child)
		if err != nil {
			return "", err
		}
		result += s
	}
	return result, nil
}

func (g *generator) generateRepeat(n *node) (string, error) {
	result := ""
	cnt := g.randomRange(1, 5)
	for i := 0; i < cnt; i++ {
		for _, child := range n.children {
			s, err := g.generateNode(child)
			if err != nil {
				return "", err
			}
			result += s
		}
	}
	return result, nil
}

func (g *generator) generateTerminalSymbol(n *node) (string, error) {
	return n.value, nil
}

func (g *generator) generateKw(n *node) (string, error) {
	return " " + n.value + " ", nil
}

var charset = "abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ _.!?0123456789ŨŪŹŕùûáéòµ¶"
var escapeSequences = []string{
	`\\`,
	`\'`,
	`\"`,
	"\\`",
	`\b`,
	`\n`,
	`\r`,
	`\f`,
	`\"`,
	`\u1E00`,
	`\UF1A4`,
}

func (g *generator) generateSingleQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(0, 5)
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

func (g *generator) generateDoubleQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(0, 5)
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

func (g *generator) generateAccentQuotedCharacterSequence(n *node) string {
	cnt := g.randomRange(0, 5)
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

func (g *generator) generateCharacterRepresentation(n *node) string {
	r, err := randChar(charset)
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *generator) generateStringLiteralCharacter(n *node) string {
	return "somerandomstring"
}

func (g *generator) generateIdentifier(n *node) (string, error) {
	if g.includeDelimitedIdentifiers {
		return g.generateAlt(n)
	}
	for _, child := range n.children {
		if child.name == "regular identifier" {
			return g.generateNode(n)
		}
	}
	panic("regular identifier not found")
}

func (g *generator) generateIdentifierStart(n *node) string {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ")
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *generator) generateIdentifierExtend(n *node) string {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ0123456789")
	if err != nil {
		panic(err)
	}
	return string(r)
}

func (g *generator) generateWhitespace(n *node) string {
	return " "
}

func (g *generator) generateTruncatingWhitespace(n *node) string {
	return " "
}

func (g *generator) generateBidirectionControlCharacter(n *node) string {
	return ""
}

func (g *generator) generateSimpleCommentCharacter(n *node) string {
	return ""
}

func (g *generator) generateBracketedCommentContents(n *node) string {
	return ""
}

func (g *generator) generateNewline(n *node) string {
	return "\n"
}

func (g *generator) generateOtherDigit(n *node) string {
	return ""
}

func (g *generator) generateOtherLanguageCharacter(n *node) string {
	return ""
}

func (g *generator) randomRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func (g *generator) randString(n int, charset string, escapeSequences []string) (string, error) {
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

func (g *generator) enterLeave(n *node) func() {
	n.cnt++
	return func() {
		n.cnt--
	}
}
