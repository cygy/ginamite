package recurring

// The name of the built-in recurring tasks.
const (
	// DeleteExpiredTokens : task to delete the expired token from the database.
	DeleteExpiredTokens = "delete_expired_tokens"

	// SendRegistrationMails : task to send mails to confirm unconfirmed accounts.
	SendRegistrationMails = "send_registration_mails"

	// DeleteNeverUsedAccounts : task to delete the never used accounts from the database, never logged in.
	DeleteNeverUsedAccounts = "delete_never_used_accounts"

	// DeleteInactiveAccounts : task to delete the inactive accounts from the database.
	DeleteInactiveAccounts = "delete_inactive_accounts"

	// GenerateSitemaps : generate the sitemaps.
	GenerateSitemaps = "generate_sitemaps"

	// CleanDeletedAccounts : clean the dleted accounts and the associated data.
	CleanDeletedAccounts = "clean_deleted_accounts"
)
