package commands

import (
	"fmt"
)

func printAll[T fmt.Stringer](infos ...T) {
	for _, info := range infos {
		fmt.Println(info)
	}
}
