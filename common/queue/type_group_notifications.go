package queue

// GroupNotification : content of a payload
type GroupNotification struct {
	Ability string `json:"ability"`
	Payload any    `json:"payload"`
}
