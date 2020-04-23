package app

import (
	"math/rand"
	"sort"
)

type Generator struct {
	entry      []weight
	entryRange float64
}

type weight struct {
	msg  byte
	slot float64
}

func NewGenerator(prct map[byte]float64) *Generator {
	g := new(Generator)

	sum := 0.0
	for k, v := range prct {
		sum += v
		g.entry = append(g.entry, weight{msg: k, slot: sum})
	}
	g.entryRange = sum

	// sort by slot ascending
	sort.Slice(g.entry, func(i, j int) bool {
		return g.entry[i].slot < g.entry[j].slot
	})

	return g
}

func (g *Generator) CreateMessage(typ byte) []byte {
	buf := make([]byte, avgSize[typ])
	rand.Read(buf)
	buf[0] = typ
	return buf
}

func (g *Generator) WeightedRandomType() byte {
	r := rand.Float64() * g.entryRange
	for _, w := range g.entry {
		if r < w.slot {
			return w.msg
		}
	}

	return g.entry[len(g.entry)-1].msg
}
