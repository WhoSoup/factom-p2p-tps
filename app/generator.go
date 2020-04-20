package app

import "math/rand"

type Generator struct {
}

func (g *Generator) CreateMessage(typ byte) []byte {
	buf := make([]byte, avgSize[typ])
	rand.Read(buf)
	return buf
}
