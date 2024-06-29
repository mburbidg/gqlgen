package main

import (
	"bufio"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestParse(t *testing.T) {
	f, err := os.Open("bnf.xml")
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)

	p := newBNFParser()
	g := p.parse(r)
	assert.NotNil(t, g)
}
