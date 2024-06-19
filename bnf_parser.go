package gqlgen

import (
	"errors"
	"io"
)

type bnfParser struct {
}

func newBNFParser() bnfParser {
	return bnfParser{}
}

func (p *bnfParser) parse(r io.Reader) (*tree, error) {
	return nil, errors.New("not implemented")
}
