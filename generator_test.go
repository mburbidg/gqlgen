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

	g := newGenerator(grammar, 6)
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

	g := newGenerator(grammar, 6)
	for i := 0; i < 1; i++ {
		s := g.generate("absolute directory path", grammar, false)
		fmt.Printf("string[%d]: %s\n", i, s)
	}
}
