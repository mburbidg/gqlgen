package main

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	f, err := os.Open("bnf.xml")
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	p := newBNFParser()
	grammar := p.parse(r)
	assert.NotNil(t, grammar)

	g := newGenerator(grammar, 6, "GQL-program", false)
	g.walk(g.grammar, func(n *node) {
		if n.kind == altKind {
			if len(n.children) > 1 {
				eq := true
				depth := n.children[0].refDepth
				for i := 1; i < len(n.children); i++ {
					if depth != n.children[i].refDepth {
						eq = false
					}
				}
				if eq {
					fmt.Printf("all children have the same refDepth: id=%d\n", n.id)
				}
			}
		}
	})
	g.printNode(g.grammar, "")
}

func TestGenerate(t *testing.T) {
	f, err := os.Open("bnf.xml")
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	p := newBNFParser()
	grammar := p.parse(r)
	assert.NotNil(t, grammar)

	g := newGenerator(grammar, 6, "GQL-program", false)
	s := g.generate("value expression", grammar, false)
	fmt.Printf("%s\n", s)
}
