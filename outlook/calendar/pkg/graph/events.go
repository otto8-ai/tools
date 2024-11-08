package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/recurrence"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type CreateEventInfo struct {
	Attendees                               []string // slice of email addresses
	Subject, Location, Body, ID, Recurrence string
	Owner                                   OwnerType
	IsOnline                                bool
	Start, End                              time.Time
}

func GetEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) (models.Eventable, error) {
	expand := []string{"attachments"}
	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			requestParameters := &users.ItemCalendarsItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: expand,
			}
			configuration := &users.ItemCalendarsItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: requestParameters,
			}
			resp, err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Get(ctx, configuration)
			if err != nil {
				return nil, fmt.Errorf("failed to get event: %w", err)
			}
			return resp, nil
		case OwnerTypeGroup:
			requestParameters := &groups.ItemEventsEventItemRequestBuilderGetQueryParameters{
				Expand: expand,
			}
			configuration := &groups.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
				QueryParameters: requestParameters,
			}
			resp, err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Get(ctx, configuration)
			if err != nil {
				return nil, fmt.Errorf("failed to get event: %w", err)
			}
			return resp, nil
		}
	}
	requestParameters := &users.ItemEventsEventItemRequestBuilderGetQueryParameters{
		Expand: expand,
	}
	configuration := &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
		QueryParameters: requestParameters,
	}
	resp, err := client.Me().Events().ByEventId(eventID).Get(ctx, configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return resp, nil
}

func CreateEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, info CreateEventInfo) (models.Eventable, error) {
	requestBody := models.NewEvent()

	if info.Recurrence != "" {
		// Recurrence is pretty complicated in the graph API, so we use an internal tool call to generate it.
		r, err := recurrence.Generate(ctx, info.Recurrence)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recurrence: %w", err)
		}

		graphRecurrence, err := r.ConvertForGraphAPI()
		if err != nil {
			return nil, fmt.Errorf("failed to convert recurrence for Graph API: %w", err)
		}

		requestBody.SetRecurrence(graphRecurrence)
	}

	var attendees []models.Attendeeable
	for _, a := range info.Attendees {
		attendee := models.NewAttendee()
		email := models.NewEmailAddress()
		email.SetAddress(&a)
		attendee.SetEmailAddress(email)
		attendees = append(attendees, attendee)
	}
	requestBody.SetAttendees(attendees)

	requestBody.SetSubject(&info.Subject)

	location := models.NewLocation()
	location.SetDisplayName(&info.Location)
	requestBody.SetLocation(location)

	body := models.NewItemBody()
	body.SetContent(&info.Body)
	body.SetContentType(util.Ptr(models.TEXT_BODYTYPE))
	requestBody.SetBody(body)

	requestBody.SetIsOnlineMeeting(&info.IsOnline)

	start := models.NewDateTimeTimeZone()
	start.SetDateTime(util.Ptr(info.Start.UTC().Format(time.RFC3339)))
	start.SetTimeZone(util.Ptr("UTC"))
	requestBody.SetStart(start)

	end := models.NewDateTimeTimeZone()
	end.SetDateTime(util.Ptr(info.End.UTC().Format(time.RFC3339)))
	end.SetTimeZone(util.Ptr("UTC"))
	requestBody.SetEnd(end)

	if info.ID != "" {
		switch info.Owner {
		case OwnerTypeUser:
			event, err := client.Me().Calendars().ByCalendarId(info.ID).Events().Post(ctx, requestBody, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create event: %w", err)
			}
			return event, nil
		case OwnerTypeGroup:
			event, err := client.Groups().ByGroupId(info.ID).Events().Post(ctx, requestBody, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create event: %w", err)
			}
			return event, nil
		default:
			return nil, fmt.Errorf("invalid owner type: %s (possible values are \"user\" and \"group\")", info.Owner)
		}
	}

	// Create the event in the user's default calendar.
	event, err := client.Me().Events().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}
	return event, nil
}

func InviteUserToEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType, userEmail, message string) error {
	requestBody := users.NewItemEventsItemForwardPostRequestBody()
	recipient := models.NewRecipient()
	email := models.NewEmailAddress()
	email.SetAddress(&userEmail)
	recipient.SetEmailAddress(email)

	requestBody.SetComment(&message)
	requestBody.SetToRecipients([]models.Recipientable{recipient})

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Forward().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to invite user to event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Forward().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to invite user to event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Forward().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to invite user to event: %w", err)
	}
	return nil
}

func DeleteEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
				return fmt.Errorf("failed to delete event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
				return fmt.Errorf("failed to delete event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Delete(ctx, nil); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}

func AcceptEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemAcceptPostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to accept event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to accept event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Accept().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to accept event: %w", err)
	}
	return nil
}

func TentativelyAcceptEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemTentativelyAcceptPostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to tentatively accept event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to tentatively accept event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).TentativelyAccept().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to tentatively accept event: %w", err)
	}
	return nil
}

func DeclineEvent(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, eventID, calendarID string, owner OwnerType) error {
	requestBody := users.NewItemEventsItemDeclinePostRequestBody()
	requestBody.SetSendResponse(util.Ptr(true))

	if calendarID != "" {
		switch owner {
		case OwnerTypeUser:
			if err := client.Me().Calendars().ByCalendarId(calendarID).Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to decline event: %w", err)
			}
			return nil
		case OwnerTypeGroup:
			if err := client.Groups().ByGroupId(calendarID).Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
				return fmt.Errorf("failed to decline event: %w", err)
			}
			return nil
		}
	}

	if err := client.Me().Events().ByEventId(eventID).Decline().Post(ctx, requestBody, nil); err != nil {
		return fmt.Errorf("failed to decline event: %w", err)
	}
	return nil
}
