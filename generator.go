package gqlgen

import (
	crand "crypto/rand"
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
	rules   map[string]*node
	grammar *node
}

func newGenerator(grammar *node) *generator {
	g := &generator{grammar: grammar}

	g.transformSeeTheRules(g.grammar)
	g.transformRepeat(g.grammar)
	g.transformAlt(g.grammar)
	g.buildRuleMap()

	//g.printNode(g.grammar, "")

	return g
}

func (g *generator) walk(n *node, visitor func(n *node)) {
	visitor(n)
	for _, child := range n.children {
		g.walk(child, visitor)
	}
}

func (g *generator) generate(w io.Writer, startRule string, tree *node) {
	g.generateNode(w, g.rules[startRule])
	io.WriteString(w, "\n")
}

func (g *generator) buildRuleMap() {
	g.rules = make(map[string]*node)
	g.walk(g.grammar, func(n *node) {
		if n.id == bnfDefId {
			g.rules[n.name] = n.children[0] // rhs
		}
	})
	g.rules["character representation"] = &node{
		id:         fnId,
		generateFn: g.generateCharacterRepresentation,
	}
	g.rules["string literal character"] = &node{
		id:         fnId,
		generateFn: g.generateStringLiteralCharacter,
	}
	g.rules["identifier start"] = &node{
		id:         fnId,
		generateFn: g.generateIdentifierStart,
	}
	g.rules["identifier extend"] = &node{
		id:         fnId,
		generateFn: g.generateIdentifierExtend,
	}
	g.rules["whitespace"] = &node{
		id:         fnId,
		generateFn: g.generateWhitespace,
	}
	g.rules["truncating whitespace"] = &node{
		id:         fnId,
		generateFn: g.generateTruncatingWhitespace,
	}
	g.rules["bidirectional control character"] = &node{
		id:         fnId,
		generateFn: g.generateBidirectionControlCharacter,
	}
	g.rules["simple comment character"] = &node{
		id:         fnId,
		generateFn: g.generateSimpleCommentCharacter,
	}
	g.rules["bracketed comment contents"] = &node{
		id:         fnId,
		generateFn: g.generateBracketedCommentContents,
	}
	g.rules["newline"] = &node{
		id:         fnId,
		generateFn: g.generateNewline,
	}
	g.rules["other digit"] = &node{
		id:         fnId,
		generateFn: g.generateOtherDigit,
	}
	g.rules["other language character"] = &node{
		id:         fnId,
		generateFn: g.generateOtherLanguageCharacter,
	}
	g.rules["single quoted character sequence"] = &node{
		id:         fnId,
		generateFn: g.generateSingleQuotedCharacterSequence,
	}
	g.rules["double quoted character sequence"] = &node{
		id:         fnId,
		generateFn: g.generateDoubleQuotedCharacterSequence,
	}
	g.rules["accent quoted character sequence"] = &node{
		id:         fnId,
		generateFn: g.generateAccentQuotedCharacterSequence,
	}
}

func (g *generator) transformRepeat(n *node) {
	var prev *node
	toRemove := []*node{}
	for _, child := range n.children {
		if child.id == repeatId {
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
		if child.id != childToRemove.id {
			c = append(c, child)
		}
	}
	n.children = c
}

func (g *generator) transformAlt(n *node) {
	for _, child := range n.children {
		g.transformAlt(child)
	}
	if len(n.children) > 1 && n.children[0].id == altId {
		alt := &node{id: altId, parent: n}
		for _, child := range n.children {
			if child.id == altId {
				if len(child.children) > 1 {
					alt.children = append(alt.children, child)
					child.id = groupId
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
		if child.id == seeTheRulesId {
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
	if n.id == kwId {
		fmt.Printf("%s%s(%s)\n", indent, n.id, n.value)
	} else if n.id == bnfId || n.id == bnfDefId {
		fmt.Printf("%s%s(%s)\n", indent, n.id, n.name)
	} else {
		fmt.Printf("%s%s\n", indent, n.id)
	}
	for _, child := range n.children {
		g.printNode(child, indent+"  ")
	}
}

func (g *generator) generateNode(w io.Writer, n *node) {
	switch n.id {
	case rhsId:
		g.generateRhs(w, n)
	case altId:
		g.generateAlt(w, n)
	case bnfId:
		g.generateBnf(w, n)
	case optId:
		g.generateOpt(w, n)
	case groupId:
		g.generateGroup(w, n)
	case repeatId:
		g.generateRepeat(w, n)
	case terminalSymbolId:
		g.generateTerminalSymbol(w, n)
	case kwId:
		g.generateKw(w, n)
	case fnId:
		n.generateFn(w, n)
	}
}

func (g *generator) generateRhs(w io.Writer, n *node) {
	if n.enterFn != nil {
		n.enterFn(w, n)
	}
	for _, child := range n.children {
		g.generateNode(w, child)
	}
	if n.leaveFn != nil {
		n.leaveFn(w, n)
	}
}

func (g *generator) generateBnf(w io.Writer, n *node) {
	if n, ok := g.rules[n.name]; ok {
		g.generateNode(w, n)
	} else {
		panic("rule not found: " + n.name)
	}
}

func (g *generator) generateAlt(w io.Writer, n *node) {
	i := g.randomRange(0, len(n.children))
	g.generateNode(w, n.children[i])
}

func (g *generator) generateOpt(w io.Writer, n *node) {
	i := g.randomRange(0, 2)
	if i == 1 {
		for _, child := range n.children {
			g.generateNode(w, child)
		}
	}
}

func (g *generator) generateGroup(w io.Writer, n *node) {
	for _, child := range n.children {
		g.generateNode(w, child)
	}
}

func (g *generator) generateRepeat(w io.Writer, n *node) {
	cnt := g.randomRange(0, 5)
	for i := 0; i < cnt; i++ {
		for _, child := range n.children {
			g.generateNode(w, child)
		}
	}
}

func (g *generator) generateTerminalSymbol(w io.Writer, n *node) {
	_, err := io.WriteString(w, n.value)
	if err != nil {
		panic(err)
	}
}

func (g *generator) generateKw(w io.Writer, n *node) {
	_, err := io.WriteString(w, " "+n.value+" ")
	if err != nil {
		panic(err)
	}
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
