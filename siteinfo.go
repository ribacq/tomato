package main

import (
	"fmt"
)

// Siteinfo contains the site-wide meta. There should be only one of them.
// Name and Title will be printed in the header,
// Description will be printed in the menu,
// Copyright will be printed in the footer.
// Authors must contain all possible authors for the website.
type Siteinfo struct {
	Name        string   `json: "name"`
	Title       string   `json: "title"`
	Subtitle    string   `json: "subtitle"`
	Description string   `json: "description"`
	Copyright   string   `json: "copyright"`
	Authors     []Author `json: "authors"`
}

// MainAuthorHelper prints a html link to the first author of the siteinfo.
func (siteinfo Siteinfo) MainAuthorHelper() string {
	return siteinfo.Authors[0].Helper()
}

// CopyrightHelper prints html for the copyright information.
func (siteinfo Siteinfo) CopyrightHelper(page Page) string {
	return string(Html([]byte(siteinfo.Copyright), page))
}

// SubtitleHelper prints html for the site subtitle.
func (siteinfo Siteinfo) SubtitleHelper(page Page) string {
	return string(Html([]byte(siteinfo.Subtitle), page))
}

// DescriptionHelper prints html for the site description.
func (siteinfo Siteinfo) DescriptionHelper(page Page) string {
	return string(Html([]byte(siteinfo.Description), page))
}

// FindAuthor returns an existing author by its name or nil and an error if there is no author with this name.
func (si *Siteinfo) FindAuthor(name string) (*Author, error) {
	for i := range si.Authors {
		if si.Authors[i].Name == name {
			return &si.Authors[i], nil
		}
	}
	return nil, fmt.Errorf("unable to find this author")
}
