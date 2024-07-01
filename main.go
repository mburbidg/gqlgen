package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	bnfIn, startRule, verbose, cnt := processArgs()
	p := newBNFParser()
	tree := p.parse(bnfIn)

	g := newGenerator(tree, 6)
	for i := 0; i < cnt; i++ {
		s := g.generate(startRule, tree, verbose)
		fmt.Printf("%s\n", s)
	}
}

func processArgs() (io.Reader, string, bool, int) {
	bnf := flag.String("bnf", "./bnf.xml", "path to an XML file containing the bnf rules for GQL")
	startRule := flag.String("start", "GQL-program", "start rule name")
	verbose := flag.Bool("v", false, "verbose")
	cnt := flag.Int("cnt", 1, "number of grammar strings")
	flag.Parse()

	f, err := os.Open(*bnf)
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	r := bufio.NewReader(f)
	return r, *startRule, *verbose, *cnt
}
