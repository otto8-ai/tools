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
