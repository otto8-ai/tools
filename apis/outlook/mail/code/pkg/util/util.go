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

func Dedupe[T any, U comparable](arr []T, f func(T) U) []T {
	seen := make(map[U]struct{})
	var out []T
	for _, v := range arr {
		key := f(v)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
