package model

import (
	"github.com/globalsign/mgo/bson"
)

/**
This struct is used only for creating the XML sitemap files.
*/

// SitemapUser : sitemap version of the User struct
type SitemapUser struct {
	ID   bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Slug string        `json:"slug" bson:"slug"`
}

// NewSitemapUser : returns a new 'SitemapUser' struct.
func NewSitemapUser(user User) SitemapUser {
	return SitemapUser{
		ID:   user.ID,
		Slug: user.Slug,
	}
}
