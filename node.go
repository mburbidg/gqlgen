package main

const (
	grammarKind        = "grammar"
	bnfDefKind         = "BNFdef"
	rhsKind            = "rhs"
	altKind            = "alt"
	bnfKind            = "BNF"
	optKind            = "opt"
	groupKind          = "group"
	repeatKind         = "repeat"
	terminalSymbolKind = "terminalsymbol"
	kwKind             = "kw"
	fnKind             = "fn"
	seeTheRulesKind    = "seeTheRules"
)

type node struct {
	id         int
	kind       string
	cnt        int
	refDepth   int
	name       string
	parent     *node
	children   []*node
	value      string
	generateFn func(n *node) string
}
