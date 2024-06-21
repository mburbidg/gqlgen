package gqlgen

import "io"

const (
	grammarId        = "grammar"
	bnfDefId         = "BNFdef"
	rhsId            = "rhs"
	altId            = "alt"
	bnfId            = "BNF"
	optId            = "opt"
	groupId          = "group"
	repeatId         = "repeat"
	terminalSymbolId = "terminalsymbol"
	kwId             = "kw"
	fnId             = "fn"
	seeTheRulesId    = "seeTheRules"
)

type node struct {
	id         string
	name       string
	parent     *node
	children   []*node
	value      string
	generateFn func(w io.Writer, n *node)
	enterFn    func(w io.Writer, n *node)
	leaveFn    func(w io.Writer, n *node)
}
