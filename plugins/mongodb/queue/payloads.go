package queue

// TaskUpdateUserPublicProfile is the payload of the MessageTaskUpdateUserPublicProfile messages.
type TaskUpdateUserPublicProfile struct {
	UserID string `json:"user_id"`
}

// TaskUpdateUserPublicImage is the payload of the MessageTaskUpdateUserPublicImage messages.
type TaskUpdateUserPublicImage struct {
	UserID string `json:"user_id"`
}
