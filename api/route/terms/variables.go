package terms

var (
	// UpdateVersionOfTermsAcceptedByUser : the function to save the latest version of the terms acepted by the user.
	UpdateVersionOfTermsAcceptedByUser UpdateVersionOfTermsAcceptedByUserFunc

	// the localized contents of the current terms.
	termsContents = map[string]string{}
)
