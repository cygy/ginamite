package html

// User holds the properties of the user of a request.
type User struct {
	ID      string
	Name    string
	Image1x string
	Image2x string
	Unread  uint
}

// Info is an information displayed by a page.
type Info struct {
	Title   string
	Message string
}

// Error is an error displayed by a page.
type Error struct {
	Message  string
	Reason   string
	Recovery string
}

// PageContent is the content of a HTML page: the properties and the variables which change with each request.
type PageContent struct {
	CurrentUser        User
	Title              string
	Error              Error
	Info               Info
	Vars               map[string]any
	parsedMainTemplate string
}

// HasTitle reports whether a title is defined.
func (content *PageContent) HasTitle() bool {
	return len(content.Title) > 0
}

// HasInfo reports whether an info message is defined.
func (content *PageContent) HasInfo() bool {
	return len(content.Info.Message) > 0
}

// HasError reports whether an error message is defined.
func (content *PageContent) HasError() bool {
	return len(content.Error.Message) > 0
}

// IsUserAuthenticated reports whether the request has a user.
func (content *PageContent) IsUserAuthenticated() bool {
	return len(content.CurrentUser.ID) > 0
}
