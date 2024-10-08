package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type CalendarInfo struct {
	Calendar models.Calendarable
	ID       string
	Owner    OwnerType
}

type OwnerType string

const (
	OwnerTypeUser  OwnerType = "user"
	OwnerTypeGroup OwnerType = "group"
)

func GetCalendar(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, owner OwnerType, id string) (CalendarInfo, error) {
	if owner == OwnerTypeUser {
		resp, err := client.Me().Calendars().ByCalendarId(id).Get(ctx, nil)
		if err != nil {
			return CalendarInfo{}, fmt.Errorf("failed to get calendar: %w", err)
		}

		return CalendarInfo{
			Calendar: resp,
			ID:       id,
			Owner:    OwnerTypeUser,
		}, nil
	}

	resp, err := client.Groups().ByGroupId(id).Calendar().Get(ctx, nil)
	if err != nil {
		return CalendarInfo{}, fmt.Errorf("failed to get calendar: %w", err)
	}

	return CalendarInfo{
		Calendar: resp,
		ID:       id,
		Owner:    OwnerTypeGroup,
	}, nil
}

func ListCalendars(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]CalendarInfo, error) {
	calendarsGetResp, err := client.Me().Calendars().Get(ctx, &users.ItemCalendarsRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarsRequestBuilderGetQueryParameters{
			Top: util.Ptr(int32(100)),
		},
	})

	// TODO - handle if there are more than 100

	if err != nil {
		return nil, fmt.Errorf("failed to list user's calendars: %w", err)
	}

	var calendars []CalendarInfo
	for _, calendar := range calendarsGetResp.GetValue() {
		calendars = append(calendars, CalendarInfo{
			Calendar: calendar,
			ID:       util.Deref(calendar.GetId()),
			Owner:    OwnerTypeUser,
		})
	}

	// Get the group memberships so that we can check group calendars.
	memberOf, err := client.Me().MemberOf().Get(ctx, &users.ItemMemberOfRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMemberOfRequestBuilderGetQueryParameters{
			Top: util.Ptr(int32(100)),
		},
	})

	// TODO - handle if there are more than 100

	if err != nil {
		return nil, fmt.Errorf("failed to get group memberships: %w", err)
	}

	for _, group := range memberOf.GetValue() {
		result, err := client.Groups().ByGroupId(util.Deref(group.GetId())).Calendar().Get(ctx, nil)
		if err != nil {
			// Some groups don't have calendars and will just error out. That's fine.
			continue
		}

		calendars = append(calendars, CalendarInfo{
			Calendar: result,
			ID:       util.Deref(group.GetId()),
			Owner:    OwnerTypeGroup,
		})
	}

	return calendars, nil
}

func ListCalendarView(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, id string, owner OwnerType, start, end *time.Time) ([]models.Eventable, error) {
	if owner == OwnerTypeUser {
		resp, err := client.Me().Calendars().ByCalendarId(id).CalendarView().Get(ctx, &users.ItemCalendarsItemCalendarViewRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemCalendarsItemCalendarViewRequestBuilderGetQueryParameters{
				EndDateTime:   util.Ptr(util.Deref(end).Format(time.RFC3339)),
				StartDateTime: util.Ptr(util.Deref(start).Format(time.RFC3339)),
				Top:           util.Ptr(int32(100)),
			},
		})

		// TODO - handle if there are more than 100

		if err != nil {
			return nil, fmt.Errorf("failed to list calendar view: %w", err)
		}

		return resp.GetValue(), nil
	}

	resp, err := client.Groups().ByGroupId(id).CalendarView().Get(ctx, &groups.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &groups.ItemCalendarViewRequestBuilderGetQueryParameters{
			EndDateTime:   util.Ptr(util.Deref(end).Format(time.RFC3339)),
			StartDateTime: util.Ptr(util.Deref(start).Format(time.RFC3339)),
			Top:           util.Ptr(int32(100)),
		},
	})

	// TODO - handle if there are more than 100

	if err != nil {
		return nil, fmt.Errorf("failed to list calendar view: %w", err)
	}

	return resp.GetValue(), nil
}
