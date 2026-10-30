package request

// Config is the configuration of a middleware.
type Config struct {
	Name                         string // The key of the URL parameter to verify.
	SkipVerifyingInvalidObjectID bool   // Accepts any value, instead of checking that it is a valid ID, see IsValidID.
	StoreKey                     string // The key which in the context the value is stored.
	MessageKey                   string // The key of the error message.
	NotFoundReasonKey            string // The key of the reason of the error "not found".
	NotFoundRecoveryKey          string // The key of the recovery of the error "not found".
	InvalidObjectIDReasonKey     string // The key of the reason of the error "invalid object ID".
	InvalidObjectIDRecoveryKey   string // The key of the recovery of the error "invalid object ID".
}
