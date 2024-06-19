package gqlgen

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	p := newBNFParser()
	g := newGenerator()

	t, err := p.parse(getInput())
	if err != nil {
		log.Fatal(err)
	}

	err = g.generate(os.Stdout, t)
	if err != nil {
		log.Fatal(err)
	}
}

func getInput() io.Reader {
	bnf := flag.String("bnf", "", "path to an XML file containing the bnf rules for GQL")
	flag.Parse()

	f, err := os.Open(*bnf)
	if err != nil {
		log.Fatalf("error opening xml: %s\n", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	return r
}
