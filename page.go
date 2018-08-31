// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/qor/i18n"
)

// Page is the representation of a single page.
// Basename is the bit that goes in the URL.
// PathToFeaturedImage should be a URL to an image that will serve as header for this page.
type Page struct {
	ID                  string
	Category            *Category
	Basename            string
	Title               string
	Authors             []*Author
	Date                string
	Tags                []string
	Draft               bool // page won’t exist at all
	Unlisted            bool // page will exist but will not appear in Recent Pages, tags or category pages
	Content             []byte
	PathToFeaturedImage string
	Locale              string
}

// NewCategoryPage creates an index page for a category.
func NewCategoryPage(cat *Category, siteinfo *Siteinfo, locales *i18n.I18n, locale string) *Page {
	return &Page{
		ID:       "index",
		Category: cat,
		Basename: "index",
		Title:    string(locales.T(locale, "categories.page_list_name", cat.Name)),
		Authors:  []*Author{&siteinfo.Authors[0]},
		Tags:     cat.Tags(locale),
		Unlisted: true,
		Content:  []byte("{{ template \"PageList\" . }}"),
		Locale:   locale,
	}
}

// NewTagPage creates a page for a tag.
func NewTagPage(tag string, tree *Category, siteinfo *Siteinfo, locales *i18n.I18n, locale string) *Page {
	return &Page{
		ID: tag,
		Category: &Category{
			Parent:   tree,
			Basename: "tag",
			Name:     "Tags",
			Pages:    map[string][]*Page{locale: tree.FilterByTag(tag, locale)},
		},
		Basename: tag,
		Title:    string(locales.T(locale, "tags.page_list_name", tag)),
		Authors:  []*Author{&siteinfo.Authors[0]},
		Tags:     []string{tag},
		Unlisted: true,
		Content:  []byte("{{ template \"PageList\" . }}"),
		Locale:   locale,
	}
}

// ContentHelper prints the page in html.
func (page *Page) ContentHelper(localePath string) string {
	return strings.Replace(string(Html(page.Content, page, localePath)), "&quot;", "\"", -1)
}

// Excerpt returns an excerpt of the beginning of the page without any html formatting.
// Its maximum length is 140 characters.
func (page Page) Excerpt() string {
	exc := string(Raw(page.Content))
	bracesRE := regexp.MustCompile("{{.*}}")
	exc = bracesRE.ReplaceAllString(exc, "")
	if len(exc) > 280 {
		exc = exc[:280] + "…"
	}
	return exc
}

// PathHelper prints the path from the root to the current page in html.
func (page Page) PathHelper(curPage Page, localePath string) string {
	var str string
	if page.Basename != "index" {
		str = fmt.Sprintf("<a href=\"%s%s\">%s</a>", curPage.PathToRoot(localePath), page.Path(), page.Title)
	}
	cat := page.Category
	for cat != nil {
		prefix := fmt.Sprintf("<a href=\"%s%sindex.html\">%s</a>", curPage.PathToRoot(localePath), cat.Path(), cat.Name)
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

// PathInLocale returns the path to the version of the page in a different locale, without the localePath.
func (page *Page) PathInLocale(locale string) string {
	// return self path if locale doesn’t change or basename is index
	if locale == page.Locale || page.Basename == "index" {
		return page.Path()
	}

	// if locale doesn’t exist, return empty string
	if _, ok := page.Category.Pages[locale]; !ok {
		return ""
	}

	// look for page in other locales
	for _, curPage := range page.Category.Pages[locale] {
		if curPage.ID == page.ID {
			return curPage.Path()
		}
	}

	// nothing found
	return ""
}

// PathToRoot returns a series of '..' in a string to give a relative path from this page to the root of the website.
func (page *Page) PathToRoot(localePath string) string {
	str := "."
	for i := 0; i < len(strings.Split(path.Join(localePath, page.Path()), "/"))-2; i++ {
		str += "/.."
	}
	return str
}
