package validator

// TokenValidator validates a auth token
type TokenValidator interface {
	SourceName() string
	VerifyTokenAndExtractInfos(token string) (TokenInfos, bool)
}
