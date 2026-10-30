package html

// PageConfiguration is the set of the configurations of the third parties used by the pages.
type PageConfiguration struct {
	Facebook       FacebookConfiguration
	Google         GoogleConfiguration
	SocialNetworks SocialNetworksConfiguration
}

// FacebookConfiguration is the configuration of the Facebook developer application.
type FacebookConfiguration struct {
	Enabled    bool
	AppID      string
	APIVersion string
}

// GoogleConfiguration is the configuration of the Google developer application.
type GoogleConfiguration struct {
	Enabled  bool
	ClientID string
}

// SocialNetworksConfiguration holds the addresses of the pages of the website on the social networks.
type SocialNetworksConfiguration struct {
	Facebook string
	Google   string
	Twitter  string
}
