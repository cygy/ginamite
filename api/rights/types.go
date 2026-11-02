package rights

import "context"

// GetUserAbilitiesFunc returns a flag to know if a user is an administrator, and the abilities of the user. The request
// fails if it returns an error.
type GetUserAbilitiesFunc func(ctx context.Context, userID string) (bool, []string, error)
