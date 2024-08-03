package main

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
