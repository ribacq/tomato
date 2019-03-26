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
	Unlisted            bool // page will exist but will not appear in Recent Pages, tags or category pages, or menus
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
		Title:    string(locales.T(locale, "categories.page_list_name", cat.Locales[locale].Name)),
		Authors:  []*Author{&siteinfo.Authors[0]},
		Tags:     cat.Tags(locale),
		Unlisted: true,
		Content:  []byte("# {{ .Page.Title }}\n{{ template \"PageList\" . }}"),
		Locale:   locale,
	}
}

// ContentHelper prints the page in html.
func (page *Page) ContentHelper(localePath string) string {
	return strings.Replace(string(Html(page.Content, page, localePath)), "&quot;", "\"", -1)
}

// Excerpt returns an excerpt of the beginning of the page without any html formatting.
// Its maximum length is 280 characters.
func (page Page) Excerpt() string {
	exc := string(Raw(page.Content))
	bracesRE := regexp.MustCompile("{{ [^{}]* }}")
	exc = bracesRE.ReplaceAllString(exc, "")
	cutLen := 280
	if len(exc) > cutLen {
		for exc[cutLen-1] != ' ' {
			cutLen--
		}
		exc = exc[:cutLen-1] + "…"
	}
	return exc
}

// PathHelper prints the path from the root to the current page in html.
func (page Page) PathHelper(curPage Page, locale, localePath string) string {
	var str string
	if page.Basename != "index" {
		str = fmt.Sprintf("<a href=\"%s\">%s</a>", path.Join(curPage.PathToRoot(localePath), localePath, page.Path()), page.Title)
	}
	cat := page.Category
	for cat != nil {
		prefix := fmt.Sprintf("<a href=\"%s\">%s</a>", path.Join(curPage.PathToRoot(localePath), localePath, cat.Path(locale), "index.html"), cat.Locales[locale].Name)
		if len(str) == 0 {
			str = prefix
		} else {
			str = prefix + " &gt; " + str
		}
		cat = cat.Parent
	}
	return str
}

// PrevNextHelper prints in html links to previous and next page in the given category.
func (page Page) PrevNextHelper(curPage Page, catPath, locale, localePath string) string {
	cat, err := curPage.Category.Tree().FindParent(path.Join("/", catPath, "catinfo.json"))
	if err != nil {
		return "plop"
	}

	pages := cat.RecentPages(-1, locale)
	var prevPage, nextPage *Page
	for i := range pages {
		if pages[i].Path() == curPage.Path() {
			if i > 0 {
				nextPage = pages[i-1]
			}
			if len(pages) > i+1 {
				prevPage = pages[i+1]
			}
			break
		}
	}

	var ret string
	if prevPage != nil {
		ret = fmt.Sprintf("<a href=\"%s\">&larr;</a>", path.Join(curPage.PathToRoot(localePath), localePath, prevPage.Path()))
	}
	if nextPage != nil {
		if ret != "" {
			ret += "&nbsp;"
		}
		ret += fmt.Sprintf("<a href=\"%s\">&rarr;</a>", path.Join(curPage.PathToRoot(localePath), localePath, nextPage.Path()))
	}

	return ret
}

// Path returns the slash-seperated path for the page, starting from the root.
func (page *Page) Path() string {
	return page.Category.Path(page.Locale) + page.Basename + ".html"
}

// PathInLocale returns the path to the version of the page in a different locale, without the localePath.
func (page *Page) PathInLocale(locale string) string {
	// return self path if locale doesn’t change
	if locale == page.Locale {
		return page.Path()
	}

	// if locale doesn’t exist, return empty string
	if _, ok := page.Category.Locales[locale]; !ok {
		return ""
	}

	// look for page in other locales
	for _, curPage := range page.Category.Locales[locale].Pages {
		if curPage.ID == page.ID && curPage.Category == page.Category {
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
