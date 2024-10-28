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
		{
			name:       "Relative Monthly NoEnd",
			recurrence: "every month on the first Thursday, beginning 2/2/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "first",
					DaysOfWeek:     []string{"Thursday"},
					Interval:       1,
					RecurrenceType: "relativeMonthly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Relative Monthly EndDate",
			recurrence: "every two months on the last Friday, from 2/2/27 to 7/7/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "last",
					DaysOfWeek:     []string{"Friday"},
					Interval:       2,
					RecurrenceType: "relativeMonthly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-02-02",
					EndDate:        "2027-07-07",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Relative Monthly Numbered",
			recurrence: "every three months on the third Tuesday, for 18 events, starting 3/3/27",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "third",
					DaysOfWeek:     []string{"Tuesday"},
					Interval:       3,
					RecurrenceType: "relativeMonthly",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-03-03",
					NumberOfOccurrences: 18,
					RecurrenceType:      "numbered",
				},
			},
		},
		{
			name:       "Absolute Yearly NoEnd",
			recurrence: "every other year on April 1st, beginning in 2027",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Month:          4,
					DayOfMonth:     1,
					Interval:       2,
					RecurrenceType: "absoluteYearly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-04-01",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Absolute Yearly EndDate",
			recurrence: "every year on April 1st, from 2027 to 2030",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Month:          4,
					DayOfMonth:     1,
					Interval:       1,
					RecurrenceType: "absoluteYearly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-04-01",
					EndDate:        "2030-04-01",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Absolute Yearly Numbered",
			recurrence: "every five years on May 4th, 9 times, starting in 2027",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Month:          5,
					DayOfMonth:     4,
					Interval:       5,
					RecurrenceType: "absoluteYearly",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-05-04",
					NumberOfOccurrences: 9,
					RecurrenceType:      "numbered",
				},
			},
		},
		{
			name:       "Relative Yearly NoEnd",
			recurrence: "every other year on the last Monday of March, beginning in 2027",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "last",
					Month:          3,
					DaysOfWeek:     []string{"Monday"},
					Interval:       2,
					RecurrenceType: "relativeYearly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-03-29",
					RecurrenceType: "noEnd",
				},
			},
		},
		{
			name:       "Relative Yearly EndDate",
			recurrence: "every year on the first Friday of March, from 2027 to 2039",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "first",
					Month:          3,
					DaysOfWeek:     []string{"Friday"},
					Interval:       1,
					RecurrenceType: "relativeYearly",
				},
				Range: RecurrenceRange{
					StartDate:      "2027-03-01",
					EndDate:        "2039-12-31",
					RecurrenceType: "endDate",
				},
			},
		},
		{
			name:       "Relative Yearly Numbered",
			recurrence: "every third year on the fourth Thursday of May, for 10 occurrences, starting in 2027",
			expected: Recurrence{
				Pattern: RecurrencePattern{
					Index:          "fourth",
					Month:          5,
					DaysOfWeek:     []string{"Thursday"},
					Interval:       3,
					RecurrenceType: "relativeYearly",
				},
				Range: RecurrenceRange{
					StartDate:           "2027-05-27",
					NumberOfOccurrences: 10,
					RecurrenceType:      "numbered",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Generate(context.Background(), tt.recurrence)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}
