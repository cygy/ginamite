package queue

// UserNotificationNewLogin : content of a payload
type UserNotificationNewLogin struct {
	EmailAddress   string `json:"email"`
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	UnsubscribeKey string `json:"unsubscribe_key"`
	IPAddress      string `json:"ip"`
	Source         string `json:"source"`
	Device         string `json:"device"`
	Locale         string `json:"locale"`
	TokenID        string `json:"token_id"`
	TokenKey       string `json:"token_key"`
}
