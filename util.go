package erikson

import "math/rand"

func Sample(n []Node) Node {
	i := rand.Intn(len(n))
	return n[i]
}

func SampleN(nodes []Node, n int) []Node {

	out := make([]Node, n)

	for i, _ := range out {
		out[i] = Sample(nodes)
	}

	return out
}
