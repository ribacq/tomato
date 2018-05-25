package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Page is the representation of a single page.
// Basename is the bit that goes in the URL.
// PathToFeaturedImage should be a URL to an image that will serve as header for this page.
type Page struct {
	Category            *Category
	Basename            string
	Title               string
	Author              *Author
	Date                string
	Tags                []string
	Draft               bool
	Content             []byte
	PathToFeaturedImage string
}

// ContentHelper prints the page in html.
func (page Page) ContentHelper() string {
	return string(Html(page.Content, page))
}

// Excerpt returns an excerpt of the beginning of the page without any html formatting.
// Its maximum length is 140 characters.
func (page Page) Excerpt() string {
	exc := string(Raw(page.Content))
	bracesRE := regexp.MustCompile("{{.*}}")
	exc = bracesRE.ReplaceAllString(exc, "")
	if len(exc) > 140 {
		exc = exc[:140] + "…"
	}
	return exc
}

// PathHelper prints the path from the root to the current page in html.
func (page Page) PathHelper(curPage Page) string {
	var str string
	if page.Basename != "index" {
		str = fmt.Sprintf("<a href=\"%s%s\">%s</a>", curPage.PathToRoot(), page.Path(), page.Title)
	}
	cat := page.Category
	for cat != nil {
		prefix := fmt.Sprintf("<a href=\"%s%sindex.html\">%s</a>", curPage.PathToRoot(), cat.Path(), cat.Name)
		if len(str) == 0 {
			str = prefix
		} else {
			str = prefix + " &gt; " + str
		}
		cat = cat.Parent
	}
	return str
}

// Path returns the slash-seperated path for the page, starting from the root.
func (page *Page) Path() string {
	return page.Category.Path() + page.Basename + ".html"
}

// PathToRoot returns a series of '../' in a string to give a relative path from this page to the root of the website.
func (page *Page) PathToRoot() string {
	str := "."
	for i := 0; i < len(strings.Split(page.Path(), "/"))-2; i++ {
		str += "/.."
	}
	return str
}
