package global

const CredentialEnv = "GPTSCRIPT_MICROSOFT_EXCEL_TOKEN"

var (
	ReadOnlyScopes = []string{"Files.Read", "User.Read"}
	AllScopes      = []string{"Files.Read", "Files.ReadWrite", "User.Read"}
)
