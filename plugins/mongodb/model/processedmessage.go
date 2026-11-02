package model

// ProcessedMessageCollection is the name of the collection of the messages the workers handled, see ProcessedMessage.
const ProcessedMessageCollection = "processedMessages"

// ProcessedMessageTTL is the number of seconds the workers remember a handled message: a message is delivered again
// within minutes, never after days.
const ProcessedMessageTTL = 7 * 24 * 3600
