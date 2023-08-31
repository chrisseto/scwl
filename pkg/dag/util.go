package dag

type Filter[T INode] func(T) bool

func filter[T INode](in []INode, predicates ...Filter[T]) []T {
	pred := func(n INode) bool {
		t, ok := n.(T)
		if !ok {
			return false
		}

		for _, p := range predicates {
			if !p(t) {
				return false
			}
		}
		return true
	}

	var out []T
	for _, n := range in {
		if pred(n) {
			out = append(out, n.(T))
		}
	}
	return out
}
