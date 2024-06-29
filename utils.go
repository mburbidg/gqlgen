package main

import (
	"crypto/rand"
	"fmt"
)

func randString(n int, charset string) (string, error) {
	s, r := make([]rune, n), []rune(charset)
	for i := range s {
		p, err := rand.Prime(rand.Reader, len(r))
		if err != nil {
			return "", fmt.Errorf("random string n %d: %w", n, err)
		}
		x, y := p.Uint64(), uint64(len(r))
		// fmt.Printf("x: %d y: %d\tx %% y = %d\trandom[%d] = %q\n", x, y, x%y, x%y, string(r[x%y]))
		s[i] = r[x%y]
	}
	return string(s), nil
}

func randChar(charset string) (rune, error) {
	str, err := randString(1, charset)
	if err != nil {
		return 0, err
	}
	return rune(str[0]), nil
}
