package gqlgen

import "io"

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
	kind       string
	name       string
	parent     *node
	children   []*node
	value      string
	generateFn func(w io.Writer, n *node)
	enterFn    func(w io.Writer, n *node)
	leaveFn    func(w io.Writer, n *node)
}
