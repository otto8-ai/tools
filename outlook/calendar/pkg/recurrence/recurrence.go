package recurrence

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

// the source for doc.md: https://learn.microsoft.com/en-us/graph/outlook-schedule-recurring-events
//
//go:embed doc.md
var doc string

type Recurrence struct {
	Pattern RecurrencePattern `json:"pattern"`
	Range   RecurrenceRange   `json:"range"`
}

type RecurrencePattern struct {
	DayOfMonth     int      `json:"dayOfMonth"`
	DaysOfWeek     []string `json:"daysOfWeek"`
	FirstDayOfWeek string   `json:"firstDayOfWeek"`
	Index          string   `json:"index"`
	Interval       int      `json:"interval"`
	Month          int      `json:"month"`
	RecurrenceType string   `json:"type"`
}

type RecurrenceRange struct {
	EndDate             string `json:"endDate"`
	StartDate           string `json:"startDate"`
	NumberOfOccurrences int    `json:"numberOfOccurrences"`
	RecurrenceType      string `json:"type"`
}

func (r Recurrence) ConvertForGraphAPI() (models.PatternedRecurrenceable, error) {
	recurrenceable := models.NewPatternedRecurrence()

	// Convert the range
	rRange := models.NewRecurrenceRange()
	if r.Range.EndDate != "" {
		endDate, err := time.Parse(time.DateOnly, r.Range.EndDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end date: %w", err)
		}
		rRange.SetEndDate(serialization.NewDateOnly(endDate))
	}
	if r.Range.StartDate != "" {
		startDate, err := time.Parse(time.DateOnly, r.Range.StartDate)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start date: %w", err)
		}
		rRange.SetStartDate(serialization.NewDateOnly(startDate))
	}
	if r.Range.NumberOfOccurrences > 0 {
		rRange.SetNumberOfOccurrences(util.Ptr(int32(r.Range.NumberOfOccurrences)))
	}
	rRange.SetTypeEscaped(util.Ptr(getRecurrenceRangeType(r.Range.RecurrenceType)))

	// Convert the pattern
	rPattern := models.NewRecurrencePattern()
	if r.Pattern.DayOfMonth > 0 {
		rPattern.SetDayOfMonth(util.Ptr(int32(r.Pattern.DayOfMonth)))
	}
	if len(r.Pattern.DaysOfWeek) > 0 {
		rPattern.SetDaysOfWeek(util.Map(r.Pattern.DaysOfWeek, getDayOfWeek))
	}
	if r.Pattern.FirstDayOfWeek != "" {
		rPattern.SetFirstDayOfWeek(util.Ptr(getDayOfWeek(r.Pattern.FirstDayOfWeek)))
	}
	if r.Pattern.Index != "" {
		rPattern.SetIndex(util.Ptr(getWeekIndex(r.Pattern.Index)))
	}
	if r.Pattern.Interval > 0 {
		rPattern.SetInterval(util.Ptr(int32(r.Pattern.Interval)))
	}
	if r.Pattern.Month > 0 {
		rPattern.SetMonth(util.Ptr(int32(r.Pattern.Month)))
	}
	rPattern.SetTypeEscaped(util.Ptr(getRecurrencePatternType(r.Pattern.RecurrenceType)))

	recurrenceable.SetPattern(rPattern)
	recurrenceable.SetRangeEscaped(rRange)
	return recurrenceable, nil
}

func getTool(description string) gptscript.ToolDef {
	return gptscript.ToolDef{
		JSONResponse: true,
		Instructions: fmt.Sprintf(`
Given the following description of a recurrence and documentation about Outlook recurrences,
output a JSON object that describes the recurrence and matches the recurrence type in the documentation.
The JSON object must include one top-level field called recurrence, which contains both a pattern and a range.

Description: %s

Documentation:

%s`, description, doc),
	}
}

func Generate(ctx context.Context, description string) (Recurrence, error) {
	g, err := gptscript.NewGPTScript()
	if err != nil {
		return Recurrence{}, err
	}

	tool := getTool(description)

	run, err := g.Evaluate(ctx, gptscript.Options{}, tool)
	if err != nil {
		return Recurrence{}, err
	}

	result, err := run.Text()
	if err != nil {
		return Recurrence{}, err
	}

	var r struct {
		Recurrence Recurrence `json:"recurrence"`
	}
	if err := json.Unmarshal([]byte(result), &r); err != nil {
		return Recurrence{}, fmt.Errorf("failed to unmarshal recurrence: %w", err)
	}

	// The Outlook Calendar API is supposed to be able to support this, but it never seems to work,
	// so we just ask the LLM to schedule multiple events if it is trying to do this.
	if len(r.Recurrence.Pattern.DaysOfWeek) > 1 {
		if r.Recurrence.Pattern.RecurrenceType == "relativeMonthly" {
			return Recurrence{}, fmt.Errorf("error: a single monthly event cannot be scheduled on multiple days of the week - please schedule separate events instead")
		} else if r.Recurrence.Pattern.RecurrenceType == "relativeYearly" {
			return Recurrence{}, fmt.Errorf("error: a single yearly event cannot be scheduled on multiple days of the week - please schedule separate events instead")
		}
	}

	return r.Recurrence, nil
}

// It's annoying that we need to have all of these conversion functions, but the ones in the Graph SDK are bad.

func getRecurrenceRangeType(t string) models.RecurrenceRangeType {
	switch strings.ToLower(t) {
	case "enddate":
		return models.ENDDATE_RECURRENCERANGETYPE
	case "numbered":
		return models.NUMBERED_RECURRENCERANGETYPE
	default:
		return models.NOEND_RECURRENCERANGETYPE
	}
}

func getRecurrencePatternType(t string) models.RecurrencePatternType {
	switch strings.ToLower(t) {
	case "daily":
		return models.DAILY_RECURRENCEPATTERNTYPE
	case "weekly":
		return models.WEEKLY_RECURRENCEPATTERNTYPE
	case "absolutemonthly":
		return models.ABSOLUTEMONTHLY_RECURRENCEPATTERNTYPE
	case "relativemonthly":
		return models.RELATIVEMONTHLY_RECURRENCEPATTERNTYPE
	case "absoluteyearly":
		return models.ABSOLUTEYEARLY_RECURRENCEPATTERNTYPE
	case "relativeyearly":
		return models.RELATIVEYEARLY_RECURRENCEPATTERNTYPE
	default:
		return models.DAILY_RECURRENCEPATTERNTYPE
	}
}

func getDayOfWeek(d string) models.DayOfWeek {
	switch strings.ToLower(d) {
	case "sunday":
		return models.SUNDAY_DAYOFWEEK
	case "monday":
		return models.MONDAY_DAYOFWEEK
	case "tuesday":
		return models.TUESDAY_DAYOFWEEK
	case "wednesday":
		return models.WEDNESDAY_DAYOFWEEK
	case "thursday":
		return models.THURSDAY_DAYOFWEEK
	case "friday":
		return models.FRIDAY_DAYOFWEEK
	case "saturday":
		return models.SATURDAY_DAYOFWEEK
	}
	return models.SUNDAY_DAYOFWEEK
}

func getWeekIndex(w string) models.WeekIndex {
	switch strings.ToLower(w) {
	case "first":
		return models.FIRST_WEEKINDEX
	case "second":
		return models.SECOND_WEEKINDEX
	case "third":
		return models.THIRD_WEEKINDEX
	case "fourth":
		return models.FOURTH_WEEKINDEX
	case "last":
		return models.LAST_WEEKINDEX
	}
	return models.FIRST_WEEKINDEX
}
