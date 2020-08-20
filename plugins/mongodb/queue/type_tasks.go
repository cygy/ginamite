package queue

// TaskUpdateUserPublicProfile : content of a payload
type TaskUpdateUserPublicProfile struct {
	UserID string `json:"user_id"`
}

// TaskUpdateUserPublicImage : content of a payload
type TaskUpdateUserPublicImage struct {
	UserID string `json:"user_id"`
}
