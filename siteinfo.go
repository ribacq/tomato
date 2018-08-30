// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"fmt"
)

// Siteinfo contains the site-wide meta. There should be only one of them.
// Title and Subtitle will be printed in the header,
// Description will be printed in the menu,
// Copyright will be printed in the footer.
// Authors must contain all possible authors for the website.
type Siteinfo struct {
	Locales map[string]struct {
		Path        string `json: "path"`
		Title       string `json: "title"`
		Subtitle    string `json: "subtitle"`
		Description string `json: "description"`
		Copyright   string `json: "copyright"`
	} `json: "locales"`
	Authors []Author `json: "authors"`
}

// MainAuthorHelper prints a html link to the first author of the siteinfo.
func (siteinfo Siteinfo) MainAuthorHelper() string {
	return siteinfo.Authors[0].Helper()
}

// Title helper prints html for the site title.
func (siteinfo Siteinfo) TitleHelper(page *Page, locale string) string {
	return siteinfo.Locales[locale].Title
}

// SubtitleHelper prints html for the site subtitle.
func (siteinfo Siteinfo) SubtitleHelper(page *Page, locale string) string {
	return string(Html([]byte(siteinfo.Locales[locale].Subtitle), page, siteinfo.Locales[locale].Path))
}

// DescriptionHelper prints html for the site description.
func (siteinfo Siteinfo) DescriptionHelper(page *Page, locale string) string {
	return string(Html([]byte(siteinfo.Locales[locale].Description), page, siteinfo.Locales[locale].Path))
}

// CopyrightHelper prints html for the copyright information.
func (siteinfo Siteinfo) CopyrightHelper(page *Page, locale string) string {
	return string(Html([]byte(siteinfo.Locales[locale].Copyright), page, siteinfo.Locales[locale].Path))
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
