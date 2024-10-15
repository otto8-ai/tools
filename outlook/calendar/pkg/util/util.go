package util

func Ptr[T any](v T) *T {
	return &v
}

func Deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}

func Map[T, U any](arr []T, f func(T) U) []U {
	var out []U
	for _, v := range arr {
		out = append(out, f(v))
	}
	return out
}

func Flatten[T any](arr [][]T) []T {
	var out []T
	for _, v := range arr {
		out = append(out, v...)
	}
	return out
}

func MapValues[T comparable, U any](m map[T]U) []U {
	// The maps.Values function returns an iterator, but we want a slice.
	var out []U
	for _, v := range m {
		out = append(out, v)
	}
	return out
}
