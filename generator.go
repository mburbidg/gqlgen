package gqlgen

import (
	"errors"
	"io"
)

type generator struct {
}

func newGenerator() generator {
	return generator{}
}

func (g generator) generate(w io.Writer, tree *tree) error {
	return errors.New("not implemented")
}
