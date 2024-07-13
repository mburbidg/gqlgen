package main

import (
	"crypto/rand"
	"fmt"
)

func randString(n int, charset string) (string, error) {
	s, r := make([]rune, n), []rune(charset)
	for i := range s {
		p, err := rand.Prime(rand.Reader, len(r))
		if err != nil {
			return "", fmt.Errorf("random string n %d: %w", n, err)
		}
		x, y := p.Uint64(), uint64(len(r))
		// fmt.Printf("x: %d y: %d\tx %% y = %d\trandom[%d] = %q\n", x, y, x%y, x%y, string(r[x%y]))
		s[i] = r[x%y]
	}
	return string(s), nil
}

func randChar(charset string) (rune, error) {
	str, err := randString(1, charset)
	if err != nil {
		return 0, err
	}
	return rune(str[0]), nil
}

func walk(n *node, visitor func(n *node)) {
	visitor(n)
	for _, child := range n.children {
		walk(child, visitor)
	}
}

func transformSeeTheRules(n *node) {
	toRemove := []*node{}
	for _, child := range n.children {
		if child.kind == seeTheRulesKind {
			toRemove = append(toRemove, child)
		} else {
			transformSeeTheRules(child)
		}
	}
	for _, child := range toRemove {
		removeChild(n, child)
	}
}

func transformRepeat(n *node) {
	var prev *node
	toRemove := []*node{}
	for _, child := range n.children {
		if child.kind == repeatKind {
			child.children = append(child.children, prev)
			prev.parent = child
			toRemove = append(toRemove, prev)
		} else {
			transformRepeat(child)
		}
		prev = child
	}
	for _, child := range toRemove {
		removeChild(n, child)
	}
}

func transformAlt(n *node) {
	for _, child := range n.children {
		transformAlt(child)
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

func assignId(root *node) {
	var nextId int
	walk(root, func(n *node) {
		n.id = nextId
		nextId++
	})
}

func nameRhs(root *node) {
	walk(root, func(n *node) {
		if n.kind == bnfDefKind {
			n.children[0].name = n.name
		}
	})
}

func removeChild(n, childToRemove *node) {
	c := []*node{}
	for _, child := range n.children {
		if child.kind != childToRemove.kind {
			c = append(c, child)
		}
	}
	n.children = c
}

func printNode(n *node, indent string) {
	if n.kind == kwKind {
		fmt.Printf("%s%s(%d, %s)\n", indent, n.kind, n.id, n.value)
	} else if n.kind == bnfKind || n.kind == bnfDefKind {
		fmt.Printf("%s%s(%d, %s)\n", indent, n.kind, n.id, n.name)
	} else {
		fmt.Printf("%s%s(%d)\n", indent, n.kind, n.id)
	}
	for _, child := range n.children {
		printNode(child, indent+"  ")
	}
}
