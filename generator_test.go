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
	f, err := os.Open("test.xml")
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	p := newBNFParser()
	grammar := p.parse(r)
	assert.NotNil(t, grammar)

	g := newGenerator(grammar, 6, "label expression")
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

	g := newGenerator(grammar, 6, "GQL-program")
	s := g.generate("GQL-program", grammar, false)
	fmt.Printf("%s\n", s)
}
