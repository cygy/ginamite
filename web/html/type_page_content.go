package html

// User struct containing the properties of a user.
type User struct {
	ID      string
	Name    string
	Image1x string
	Image2x string
	Unread  uint
}

// Info struct containing the properties of an info.
type Info struct {
	Title   string
	Message string
}

// Error struct containing the properties of an error.
type Error struct {
	Message  string
	Reason   string
	Recovery string
}

// PageContent struct of the content of a HTML page containing some variables.
type PageContent struct {
	CurrentUser        User
	Title              string
	Error              Error
	Info               Info
	Vars               map[string]interface{}
	parsedMainTemplate string
}

// HasTitle : returns true if a title is defined.
func (content *PageContent) HasTitle() bool {
	return len(content.Title) > 0
}

// HasInfo : returns true if an info is defined.
func (content *PageContent) HasInfo() bool {
	return len(content.Info.Message) > 0
}

// HasError : returns true if an error is defined.
func (content *PageContent) HasError() bool {
	return len(content.Error.Message) > 0
}

// IsUserAuthenticated : returns true if the userID is defined.
func (content *PageContent) IsUserAuthenticated() bool {
	return len(content.CurrentUser.ID) > 0
}
