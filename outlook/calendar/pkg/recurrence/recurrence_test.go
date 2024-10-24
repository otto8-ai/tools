package recurrence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecurrenceGeneration(t *testing.T) {
	tests := []struct {
		name       string
		recurrence string
		expected   Recurrence
	}{
		{
			name:       "Daily NoEnd",
			recurrence: "every 3 days, beginning 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Interval:       3,
					RecurrenceType: "daily",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Daily EndDate",
			recurrence: "every 4 days, from 2/2/27 to 4/4/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Interval:       4,
					RecurrenceType: "daily",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					EndDate:        "2027-04-04",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Daily Numbered",
			recurrence: "every 5 days, for 10 occurrences, starting 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Interval:       5,
					RecurrenceType: "daily",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-02-02",
					NumberOfOccurrences: 10,
					RecurrenceType:      "numbered",
				},
			},
		},
		{
			name:       "Weekly NoEnd",
			recurrence: "every other week on Monday, beginning 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DaysOfWeek:     []string{"Monday"},
					Interval:       2,
					RecurrenceType: "weekly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Weekly EndDate",
			recurrence: "every week on Tuesday, from 2/2/27 to 4/4/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DaysOfWeek:     []string{"Tuesday"},
					RecurrenceType: "weekly",
					Interval:       1,
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					EndDate:        "2027-04-04",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Weekly Numbered",
			recurrence: "every three weeks on Wednesday, Thursday, and Friday, 5 times, starting 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DaysOfWeek:     []string{"Wednesday", "Thursday", "Friday"},
					Interval:       3,
					RecurrenceType: "weekly",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-02-02",
					NumberOfOccurrences: 5,
					RecurrenceType:      "numbered",
				},
			},
		},
		{
			name:       "Absolute Monthly NoEnd",
			recurrence: "every month on the 15th, beginning 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DayOfMonth:     15,
					Interval:       1,
					RecurrenceType: "absoluteMonthly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Absolute Monthly EndDate",
			recurrence: "every other month on the 15th, from 2/2/27 to 4/4/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DayOfMonth:     15,
					Interval:       2,
					RecurrenceType: "absoluteMonthly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					EndDate:        "2027-04-04",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Absolute Monthly Numbered",
			recurrence: "every three months on the 16th, 9 times, starting 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					DayOfMonth:     16,
					Interval:       3,
					RecurrenceType: "absoluteMonthly",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-02-02",
					NumberOfOccurrences: 9,
					RecurrenceType:      "numbered",
				},
			},
		},
		// TODO - add more tests
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Generate(context.Background(), tt.recurrence)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}
