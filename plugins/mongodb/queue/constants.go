// Package queue defines the messages the plugin sends to the workers, and their payloads.
package queue

const (
	// MessageTaskUpdateUserPublicProfile asks the workers to update the documents that use the public profile of a user.
	// Its payload is a TaskUpdateUserPublicProfile.
	MessageTaskUpdateUserPublicProfile = "update_user_public_profile"

	// MessageTaskUpdateUserPublicImage asks the workers to update the documents that use the public image of a user.
	// Its payload is a TaskUpdateUserPublicImage.
	MessageTaskUpdateUserPublicImage = "update_user_public_image"
)
