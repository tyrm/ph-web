package models

import (
	"fmt"

	"github.com/lib/pq"
)

type Client struct {
	Name string

	ClientID string
	Secret   string
	Domain   string
	User     *User

	CreatedAt pq.NullTime
	UpdatedAt pq.NullTime
	DeletedAt pq.NullTime
}

// GetID client id
func (c *Client) GetID() string {
	return c.ClientID
}

// GetSecret client domain
func (c *Client) GetSecret() string {
	return c.Secret
}

// GetDomain client domain
func (c *Client) GetDomain() string {
	return c.Domain
}

// GetUserID user id
func (c *Client) GetUserID() string {
	return fmt.Sprint(c.User.Token)
}
