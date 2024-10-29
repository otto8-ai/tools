package env

import (
	"os"
	"strconv"
)

func GetIntFromEnvOrDefault(env string, def int) int {
	v, _ := strconv.Atoi(os.Getenv(env))
	if v != 0 {
		return v
	}

	return def
}
