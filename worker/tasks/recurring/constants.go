// Package recurring schedules the recurring tasks of the worker.
package recurring

// The names of the built-in recurring tasks.
const (
	// DeleteExpiredTokens is the task deleting the expired auth tokens from the database.
	DeleteExpiredTokens = "delete_expired_tokens"

	// DeleteNeverUsedAccounts is the task deleting the never used accounts, never logged in, from the database.
	DeleteNeverUsedAccounts = "delete_never_used_accounts"

	// DeleteInactiveAccounts is the task deleting the inactive accounts from the database.
	DeleteInactiveAccounts = "delete_inactive_accounts"

	// GenerateSitemaps is the task generating the sitemaps.
	GenerateSitemaps = "generate_sitemaps"

	// SanitizeAccounts is the task sanitizing the deleted accounts and their associated data.
	SanitizeAccounts = "sanitize_accounts"
)
