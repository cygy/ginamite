package queue

// TaskIPLocation : content of a payload
type TaskIPLocation struct {
	IPAddress string `json:"ip"`
	TokenID   string `json:"token_id"`
}

// TaskDeleteUser : content of a payload
type TaskDeleteUser struct {
	UserID string `json:"user_id"`
}

// TaskDisableUser : content of a payload
type TaskDisableUser struct {
	UserID string `json:"user_id"`
}

// TaskEnableUser : content of a payload
type TaskEnableUser struct {
	UserID string `json:"user_id"`
}

// TaskUpdateUserSocialNetworks : content of a payload
type TaskUpdateUserSocialNetworks struct {
	UserID string `json:"user_id"`
}
