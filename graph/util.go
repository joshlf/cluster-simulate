package graph

func makeIntFunc(n int) func() int {
	return func() int {
		return n
	}
}
