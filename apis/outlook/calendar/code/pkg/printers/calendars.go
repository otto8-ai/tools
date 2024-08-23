package printers

import (
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/util"
)

func PrintCalendar(calendar graph.CalendarInfo) {
	fmt.Printf("Name: %s\n", util.Deref(calendar.Calendar.GetName()))
	fmt.Printf("  ID: %s\n", calendar.ID)
	fmt.Printf("  Owner: %s (%s)\n", util.Deref(calendar.Calendar.GetOwner().GetName()), util.Deref(calendar.Calendar.GetOwner().GetAddress()))
	fmt.Printf("  Owner Type: %s\n", string(calendar.Owner))
	fmt.Println()
}

func PrintCalendars(calendars []graph.CalendarInfo) {
	for _, c := range calendars {
		PrintCalendar(c)
	}
}
