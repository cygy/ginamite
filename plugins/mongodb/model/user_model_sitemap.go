package model

import "go.mongodb.org/mongo-driver/v2/bson"

/**
This struct is used only for creating the XML sitemap files.
*/

// SitemapUser : sitemap version of the User struct
type SitemapUser struct {
	ID   bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Slug string        `json:"slug" bson:"slug"`
}

// NewSitemapUser : returns a new 'SitemapUser' struct.
func NewSitemapUser(user User) SitemapUser {
	return SitemapUser{
		ID:   user.ID,
		Slug: user.Slug,
	}
}
