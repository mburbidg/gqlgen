package main

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestNewKpathGenerator(t *testing.T) {
	f, err := os.Open("expr.xml")
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	p := newBNFParser()
	grammar := p.parse(r)
	assert.NotNil(t, grammar)

	g := newKpathGenerator(grammar)
	g.generate(5)
}
