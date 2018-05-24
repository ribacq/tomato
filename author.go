package main

import (
	"fmt"
)

// Author is the type for an author of the website.
type Author struct {
	Name  string `json: "name"`
	Email string `json: "email"`
}

// Helper prints a html link to an author.
func (author *Author) Helper() string {
	return fmt.Sprintf("<address><a href=\"mailto:%s\">%s</a></address>", author.Email, author.Name)
}
