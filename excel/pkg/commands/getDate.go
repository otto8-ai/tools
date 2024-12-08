package commands

import (
	"fmt"
	"time"
)

func GetDate(serials []int) {
	startDate := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, v := range serials {
		date := startDate.AddDate(0, 0, v-2).Format(time.DateOnly)
		fmt.Printf("%d = %s\n", v, date)
	}
}
