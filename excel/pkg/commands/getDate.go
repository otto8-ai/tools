package commands

import (
	"fmt"
	"time"
)

func GetDate(serials []int) {
	startDate := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, v := range serials {
		// Excel treats the year 1900 as a leap year because of Lotus 1-2-3, and it starts with 1 representing January 1st, 1900.
		// To convert the Excel serial number to a date, we must therefore subtract 2 (days) from the serial number to account for those 2 days.
		// https://learn.microsoft.com/en-us/office/troubleshoot/excel/wrongly-assumes-1900-is-leap-year
		date := startDate.AddDate(0, 0, v-2).Format(time.DateOnly)
		fmt.Printf("%d = %s\n", v, date)
	}
}
