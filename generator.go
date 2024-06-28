package gqlgen

import (
	"bytes"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/rand"
)

const (
	singleQuotedCharacterSequence = iota
	doubleQuotedCharacterSequence
	accentQuotedCharacterSequence
)

type generator struct {
	rules     map[string]*node
	grammar   *node
	maxRevist int
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

func (g *generator) generate(w io.Writer, startRule string, tree *node) {
	if start, ok := g.rules[startRule]; ok {
		for {
			buf := bytes.NewBufferString("")
			err := g.generateNode(buf, start)
			if err == nil {
				str := buf.String()
				io.WriteString(w, str)
				io.WriteString(w, "\n")
				return
			} else {
				//fmt.Printf("%s\n", err)
			}
			// Todo: Before retry, we need to reset the cnt property of all nodes in the grammar.
			g.resetCnts(g.grammar)
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

func (g *generator) generateNode(w io.Writer, n *node) error {
	if n.cnt > g.maxRevist {
		return errRecursionLevelExceeded
	}
	switch n.kind {
	case altKind:
		return g.generateAlt(w, n)
	case bnfKind:
		return g.generateBnf(w, n)
	case optKind:
		return g.generateOpt(w, n)
	case groupKind:
		return g.generateGroup(w, n)
	case repeatKind:
		return g.generateRepeat(w, n)
	case terminalSymbolKind:
		return g.generateTerminalSymbol(w, n)
	case kwKind:
		return g.generateKw(w, n)
	case fnKind:
		n.generateFn(w, n)
	}
	return nil
}

func (g *generator) generateBnf(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	if n, ok := g.rules[n.name]; ok {
		return g.generateNode(w, n)
	} else {
		panic("rule not found: " + n.name)
	}
	return nil
}

func (g *generator) generateAlt(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	i := g.randomRange(0, len(n.children))
	return g.generateNode(w, n.children[i])
}

func (g *generator) generateOpt(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	i := g.randomRange(0, 2)
	if i == 1 {
		for _, child := range n.children {
			err := g.generateNode(w, child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) generateGroup(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	for _, child := range n.children {
		err := g.generateNode(w, child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *generator) generateRepeat(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	cnt := g.randomRange(0, 5)
	for i := 0; i < cnt; i++ {
		for _, child := range n.children {
			err := g.generateNode(w, child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *generator) generateTerminalSymbol(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	_, err := io.WriteString(w, n.value)
	if err != nil {
		panic(err)
	}
	return nil
}

func (g *generator) generateKw(w io.Writer, n *node) error {
	defer g.enterLeave(w, n)
	_, err := io.WriteString(w, " "+n.value+" ")
	if err != nil {
		panic(err)
	}
	return nil
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

func (g *generator) generateSingleQuotedCharacterSequence(w io.Writer, n *node) {
	cnt := g.randomRange(0, 50)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"\"`", []string{"''"})
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "@\""+guts+"\"")
		if err != nil {
			panic(err)
		}

	} else {
		guts, err := g.randString(cnt, charset+"\"`", append(escapeSequences, "''"))
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "'"+guts+"'")
		if err != nil {
			panic(err)
		}
	}
}

func (g *generator) generateDoubleQuotedCharacterSequence(w io.Writer, n *node) {
	cnt := g.randomRange(0, 50)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"'`", []string{`""`})
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "@\""+guts+"\"")
		if err != nil {
			panic(err)
		}

	} else {
		guts, err := g.randString(cnt, charset+"'`", append(escapeSequences, `""`))
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "\""+guts+"\"")
		if err != nil {
			panic(err)
		}
	}
}

func (g *generator) generateAccentQuotedCharacterSequence(w io.Writer, n *node) {
	cnt := g.randomRange(0, 50)
	if g.randomRange(0, 100) > 90 {
		// Use escaping only 10% of the time.
		guts, err := g.randString(cnt, charset+"\"'", []string{"``"})
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "@`"+guts+"`")
		if err != nil {
			panic(err)
		}

	} else {
		guts, err := g.randString(cnt, charset+"\"'", append(escapeSequences, "``"))
		if err != nil {
			panic(err)
		}
		_, err = io.WriteString(w, "`"+guts+"`")
		if err != nil {
			panic(err)
		}
	}
}

func (g *generator) generateCharacterRepresentation(w io.Writer, n *node) {
	r, err := randChar(charset)
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, string(r))
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateStringLiteralCharacter(w io.Writer, n *node) {
	_, err := io.WriteString(w, "somerandomstring")
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateIdentifierStart(w io.Writer, n *node) {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ")
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, string(r))
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateIdentifierExtend(w io.Writer, n *node) {
	r, err := randChar("_abcdefghijklmnopqursuvwxyzABCDEFGHIJKLMONPQUSTUVWXYZ0123456789")
	if err != nil {
		panic(err)
	}
	_, err = io.WriteString(w, string(r))
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateWhitespace(w io.Writer, n *node) {
	_, err := io.WriteString(w, " ")
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateTruncatingWhitespace(w io.Writer, n *node) {
	_, err := io.WriteString(w, " ")
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateBidirectionControlCharacter(w io.Writer, n *node) {
}

func (g *generator) generateSimpleCommentCharacter(w io.Writer, n *node) {
}

func (g *generator) generateBracketedCommentContents(w io.Writer, n *node) {
}

func (g *generator) generateNewline(w io.Writer, n *node) {
	_, err := io.WriteString(w, "\n")
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateOtherDigit(w io.Writer, n *node) {
}

func (g *generator) generateOtherLanguageCharacter(w io.Writer, n *node) {
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

func (g *generator) enterLeave(w io.Writer, n *node) func() {
	n.cnt++
	if n.enterFn != nil {
		n.enterFn(w, n)
	}
	return func() {
		if n.leaveFn != nil {
			n.leaveFn(w, n)
		}
		n.cnt--
	}
}
