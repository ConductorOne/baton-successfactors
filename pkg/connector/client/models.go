package client

import (
	"encoding/json"
	"time"
)

type BearerToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ErrorResponse struct {
	Schemas  []string `json:"schemas,omitempty"`
	ScimType string   `json:"scimType,omitempty"`
	Detail   string   `json:"detail,omitempty"`
	Status   int      `json:"status,omitempty"`
}

func (er *ErrorResponse) Message() string {
	if er.Detail == "" {
		return "Error response empty"
	}
	return er.Detail
}

type Response struct {
	Schemas      []string        `json:"schemas"`
	TotalResults int             `json:"totalResults"`
	ItemsPerPage int             `json:"itemsPerPage"`
	StartIndex   int             `json:"startIndex"`
	Resources    json.RawMessage `json:"resources"`
}

type User struct {
	Schemas           []string      `json:"schemas"`
	Meta              *UserMetadata `json:"meta,omitempty"`
	Id                string        `json:"id"`
	UserName          string        `json:"userName"`
	Name              Name          `json:"name"`
	Locale            string        `json:"locale"`
	UserType          string        `json:"userType"`
	Active            bool          `json:"active"`
	DisplayName       string        `json:"displayName"`
	PreferredLanguage string        `json:"preferredLanguage"`
	Emails            []struct {
		Type    string `json:"type"`
		Value   string `json:"value"`
		Primary bool   `json:"primary"`
	} `json:"emails"`
	Title string `json:"title"`
}

type Name struct {
	Formatted       string `json:"formatted"`
	FamilyName      string `json:"familyName"`
	GivenName       string `json:"givenName"`
	HonorificPrefix string `json:"honorificPrefix"`
}

type UserMetadata struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location"`
	Version      string    `json:"version"`
}

type Group struct {
	Schemas     []string       `json:"schemas"`
	Id          string         `json:"id"`
	DisplayName string         `json:"displayName"`
	Members     []*GroupMember `json:"members"`
	Meta        *GroupMetadata `json:"meta,omitempty"`
}

type GroupMember struct {
	Value string `json:"value"`
	Type  string `json:"type"`
	Ref   string `json:"$ref"`
}

type GroupMetadata struct {
	ResourceType string    `json:"resourceType"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
	Location     string    `json:"location"`
	Version      string    `json:"version"`
	MembersCnt   int       `json:"members.cnt"`
}
