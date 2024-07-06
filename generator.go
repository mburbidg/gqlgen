package main

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strings"
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
	debug     bool
}

var errRecursionLevelExceeded = errors.New("recursion level exceeded")

func newGenerator(grammar *node, maxRevisit int, rootRule string, debug bool) *generator {
	g := &generator{grammar: grammar, maxRevist: maxRevisit, debug: debug}

	g.transformSeeTheRules(g.grammar)
	g.transformRepeat(g.grammar)
	g.transformAlt(g.grammar)
	g.condenseRhs(g.grammar)
	g.assignId(g.grammar)
	g.buildRuleMap()
	g.calcRefDepths(rootRule)

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
			s, err := g.generateNode(start, false)
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
		fmt.Printf("%s%s(%d, %s) refDepth=%d\n", indent, n.kind, n.id, n.name, n.refDepth)
	} else {
		fmt.Printf("%s%s(%d) refDepth=%d\n", indent, n.kind, n.id, n.refDepth)
	}
	for _, child := range n.children {
		g.printNode(child, indent+"  ")
	}
}

func (g *generator) print(n *node, indent string) {
	switch n.kind {
	case kwKind, terminalSymbolKind:
		fmt.Printf("%s%s(%d, %s) {\n", indent, n.kind, n.id, n.value)
	case bnfKind:
		fmt.Printf("%s%s(%d, %s) {\n", indent, n.kind, n.id, n.name)
	default:
		fmt.Printf("%s%s(%d) {\n", indent, n.kind, n.id)
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

func (g *generator) calcRefDepths(rootName string) {
	leafRules := map[*node]string{}

	// First find all the leaf rules.
	for k, v := range g.rules {
		hasRef := false
		g.walk(v, func(n *node) {
			if n.kind == bnfKind {
				hasRef = true
			}
		})
		if !hasRef {
			leafRules[v] = k
		}
	}

	root, ok := g.rules[rootName]
	if !ok {
		panic(fmt.Sprintf("no rule found for %s", rootName))
	}
	indent := ""
	g.calculateRefDepths(rootName, root, leafRules, map[int]bool{}, &indent)

	// Uncomment to fixup infinity reference depths.
	//for {
	//	if changes := g.fixupRefDepths(); changes == 0 {
	//		break
	//	}
	//}
}

func (g *generator) fixupRefDepths() int {
	changes := 0

	// Fixup BNF nodes so that they are always 1 greater than the rule they reference.
	g.walk(g.grammar, func(n *node) {
		oldDepth := n.refDepth
		if n.kind == bnfKind {
			if next, ok := g.rules[n.name]; ok {
				n.refDepth = g.incRefDepth(next.refDepth, 1)
			} else {
				panic(fmt.Sprintf("no rule found for %s", n.name))
			}
		}
		if n.refDepth != oldDepth {
			changes++
		}
	})

	// Fixup other rules to account for the previous fixup.
	g.walk(g.grammar, func(n *node) {
		oldDepth := n.refDepth
		switch n.kind {
		case altKind:
			mn := math.MaxInt
			for _, child := range n.children {
				if child.refDepth < mn {
					mn = child.refDepth
				}
			}
			n.refDepth = mn
		case groupKind, repeatKind, optKind:
			mx := 0
			for _, child := range n.children {
				if child.refDepth > mx {
					mx = child.refDepth
				}
			}
			n.refDepth = mx
		}
		if n.refDepth != oldDepth {
			changes++
		}
	})

	return changes
}

func (g *generator) calculateRefDepths(name string, n *node, leafRules map[*node]string, seen map[int]bool, indent *string) int {
	// Print debugging information
	if g.debug {
		g.print(n, *indent)
		*indent = *indent + "  "
		defer func(n *node) {
			*indent = strings.TrimPrefix(*indent, "  ")
			fmt.Printf("%s}\n", *indent)
		}(n)
	}

	// Check to see if we've seen the node n, in the call stack. If so return infinity.
	if _, ok := seen[n.id]; ok {
		if g.debug {
			fmt.Printf("%sstopped\n", *indent)
		}
		n.refDepth = math.MaxInt
		return n.refDepth
	}

	// Check to see if refDepth has been calculated for the node n. If so, return it.
	if n.refDepth > 0 {
		if g.debug {
			fmt.Printf("%struncated\n", *indent)
		}
		return n.refDepth
	}

	// Add the node n to the seen set. Remove it on exit.
	seen[n.id] = true
	defer delete(seen, n.id)

	// Check to see if the node n is a leaf node. If so, return 0.
	if _, ok := leafRules[n]; ok {
		return 0
	}

	// Calculate, recursively the refDepth for the node n.
	switch n.kind {
	case bnfKind:
		if next, ok := g.rules[n.name]; ok {
			n.refDepth = g.incRefDepth(g.calculateRefDepths(n.name, next, leafRules, seen, indent), 1)
		} else {
			panic(fmt.Sprintf("no rule found for %s", n.name))
		}
	case groupKind, repeatKind, optKind:
		mx := 0
		for _, child := range n.children {
			depth := g.calculateRefDepths(name, child, leafRules, seen, indent)
			if depth > mx {
				mx = depth
			}
		}
		n.refDepth = mx
	case altKind:
		mn := math.MaxInt
		for _, child := range n.children {
			depth := g.calculateRefDepths(name, child, leafRules, seen, indent)
			if depth < mn {
				mn = depth
			}
		}
		n.refDepth = mn
	case kwKind, terminalSymbolKind:
	default:
		if g.debug {
			fmt.Printf("'%d' not handled: kind=%s\n", n.id, n.kind)
		}
	}
	return n.refDepth
}

func (g *generator) incRefDepth(depth int, by int) int {
	if depth != math.MaxInt {
		return depth + by
	}
	return depth
}

func (g *generator) getRule(rules map[*node]string) (*node, string) {
	for k, v := range rules {
		delete(rules, k)
		return k, v
	}
	return nil, ""
}

func (g *generator) generateNode(n *node, shortestPath bool) (string, error) {
	defer g.enterLeave(n)()
	if n.cnt > g.maxRevist {
		shortestPath = true
	}
	//fmt.Printf("generate node: kind=%s, id=%d\n", n.kind, n.id)
	switch n.kind {
	case altKind:
		return g.generateAlt(n, shortestPath)
	case bnfKind:
		return g.generateBnf(n, shortestPath)
	case optKind:
		return g.generateOpt(n, shortestPath)
	case groupKind:
		return g.generateGroup(n, shortestPath)
	case repeatKind:
		return g.generateRepeat(n, shortestPath)
	case terminalSymbolKind:
		return g.generateTerminalSymbol(n, shortestPath)
	case kwKind:
		return g.generateKw(n, shortestPath)
	case fnKind:
		return n.generateFn(n), nil
	}
	return "", nil
}

func (g *generator) generateBnf(n *node, shortestPath bool) (string, error) {
	if n, ok := g.rules[n.name]; ok {
		return g.generateNode(n, shortestPath)
	} else {
		panic("rule not found: " + n.name)
	}
	return "", nil
}

func (g *generator) generateAlt(n *node, shortestPath bool) (string, error) {
	if shortestPath {
		i := g.getShortestPath(n)
		child := n.children[i]
		fmt.Printf("Shortest path: altId=%d, id=%d, refDepth=%d\n", n.id, child.id, child.refDepth)
		return g.generateNode(n.children[i], shortestPath)
	} else {
		i := g.randomRange(0, len(n.children))
		return g.generateNode(n.children[i], shortestPath)
	}
	//fmt.Printf("alt id=%d, i=%d, child=%d\n", n.id, i, n.children[0].id)
}

func (g *generator) getShortestPath(n *node) int {
	j := 0
	mn := math.MaxInt
	for i, child := range n.children {
		if child.refDepth < mn {
			mn = child.refDepth
			j = i
		}
	}
	return j
}

func (g *generator) getTerminatingPath(n *node) int {
	for {
		i := g.randomRange(0, len(n.children))
		if n.children[i].refDepth != math.MaxInt {
			return i
		}
	}
}

func (g *generator) generateOpt(n *node, shortestPath bool) (string, error) {
	result := ""
	i := g.randomRange(0, 2)
	if i == 1 {
		for _, child := range n.children {
			s, err := g.generateNode(child, shortestPath)
			if err != nil {
				return "", err
			}
			result += s
		}
	}
	return result, nil
}

func (g *generator) generateGroup(n *node, shortestPath bool) (string, error) {
	result := ""
	for _, child := range n.children {
		s, err := g.generateNode(child, shortestPath)
		if err != nil {
			return "", err
		}
		result += s
	}
	return result, nil
}

func (g *generator) generateRepeat(n *node, shortestPath bool) (string, error) {
	result := ""
	cnt := g.randomRange(0, 5)
	for i := 0; i < cnt; i++ {
		for _, child := range n.children {
			s, err := g.generateNode(child, shortestPath)
			if err != nil {
				return "", err
			}
			result += s
		}
	}
	return result, nil
}

func (g *generator) generateTerminalSymbol(n *node, shortestPath bool) (string, error) {
	return n.value, nil
}

func (g *generator) generateKw(n *node, shortestPath bool) (string, error) {
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
	cnt := g.randomRange(0, 50)
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
	cnt := g.randomRange(0, 50)
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
	cnt := g.randomRange(0, 50)
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
