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

// Merge merges two maps where the values are arrays of the same type.
// The resulting map will contain all keys from both maps, with each value containing the concatenated arrays from both maps.
func Merge[T comparable, U any](one map[T][]U, two map[T][]U) map[T][]U {
	out := make(map[T][]U, len(one) + len(two))
	checked := make(map[T]struct{})

	for k, v := range one {
		checked[k] = struct{}{}
		out[k] = v
		if _, ok := two[k]; ok {
			out[k] = append(out[k], two[k]...)
		}
	}

	for k, v := range two {
		if _, ok := checked[k]; !ok {
			out[k] = v
		}
	}

	return out
}
