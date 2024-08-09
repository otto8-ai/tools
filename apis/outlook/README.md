# Outlook API

These tools are for interacting with the Microsoft Graph API for Outlook to manage email and calendar.
The GPTScript Gateway is required to facilitate the OAuth authorization flow for these tools.

## Mail

- github.com/gptscript-ai/tools/apis/outlook/mail/read
  - Contains tools to read emails, but cannot modify, create, or delete anything.
- github.com/gptscript-ai/integrations/outlook/mail/manage
  - Contains tools to read, modify, create, and delete emails.

## Calendar

- github.com/gptscript-ai/tools/apis/outlook/calendar/read
  - Contains tools to find information in your calendars, but not create any new events.
- github.com/gptscript-ai/tools/apis/outlook/calendar/manage
  - Contains tools to create new events in your calendars, respond to invites, invite people to your events, and delete events.
