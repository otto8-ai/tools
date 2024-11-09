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

func Filter[T any](arr []T, f func(T) bool) []T {
	var out []T
	for _, v := range arr {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}
