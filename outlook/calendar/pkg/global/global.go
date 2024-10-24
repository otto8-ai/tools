package global

const CredentialEnv = "GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN"

var (
	ReadOnlyScopes = []string{"Calendars.Read", "Calendars.Read.Shared", "Group.Read.All", "GroupMember.Read.All", "User.Read", "MailboxSettings.Read"}
	AllScopes      = []string{"Calendars.Read", "Calendars.Read.Shared", "Calendars.ReadWrite", "Calendars.ReadWrite.Shared", "Group.Read.All", "Group.ReadWrite.All", "GroupMember.Read.All", "User.Read"}
)
