package tokengen

import (
	"crypto/rand"
	"encoding/hex"
)

type Generator interface {
	Generate() string
}

type randGenerator struct {
	tokenLength int
}

func NewRandGenerator(tokenLength int) Generator {
	return &randGenerator{tokenLength: tokenLength}
}

func (g *randGenerator) Generate() string {
	b := make([]byte, g.tokenLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
