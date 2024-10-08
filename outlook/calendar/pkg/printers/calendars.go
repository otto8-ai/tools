package printers

import (
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
)

func PrintCalendar(calendar graph.CalendarInfo) {
	fmt.Printf("Name: %s\n", util.Deref(calendar.Calendar.GetName()))
	fmt.Printf("  ID: %s\n", calendar.ID)
	if calendar.Calendar.GetOwner() != nil {
		fmt.Printf("  Owner: %s (%s)\n", util.Deref(calendar.Calendar.GetOwner().GetName()), util.Deref(calendar.Calendar.GetOwner().GetAddress()))
		fmt.Printf("  Owner Type: %s\n", string(calendar.Owner))
	} else {
		fmt.Printf("  Owner: unknown\n")
		fmt.Printf("  Owner Type: unknown\n")
	}
	fmt.Println()
}

func PrintCalendars(calendars []graph.CalendarInfo) {
	for _, c := range calendars {
		PrintCalendar(c)
	}
}
