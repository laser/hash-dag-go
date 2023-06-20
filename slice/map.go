package slice

func Map[T any, U any](xs []T, f func(T) U) []U {
	out := make([]U, 0, len(xs))

	for _, v := range xs {
		out = append(out, f(v))
	}

	return out
}
