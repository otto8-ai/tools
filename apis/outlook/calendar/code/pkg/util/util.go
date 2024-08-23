package util

import "time"

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

func Dedupe[T any, U comparable](arr []T, id func(T) U) []T {
	seen := make(map[U]bool)
	var out []T
	for _, v := range arr {
		i := id(v)
		if !seen[i] {
			seen[i] = true
			out = append(out, v)
		}
	}
	return out
}

func TimeFilter(start, end *time.Time) string {
	if start != nil && end != nil {
		return "start/dateTime ge " + start.Format(time.RFC3339) + " and end/dateTime le " + end.Format(time.RFC3339)
	} else if start != nil {
		return "start/dateTime ge " + start.Format(time.RFC3339)
	} else if end != nil {
		return "end/dateTime le " + end.Format(time.RFC3339)
	}
	return ""
}
