// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

// Category represents a category, that is, a directory in the tree.
// Name and Description are fetched from a `catinfo.json` file that should
// Basename is the bit that goes in the URL.
// be located at the root of every directory.
type Category struct {
	Parent        *Category          `json: "-"`
	Basename      string             `json: "-"`
	Name          string             `json: "name"`
	Description   string             `json: "description"`
	SubCategories []*Category        `json: "-"`
	Pages         map[string][]*Page `json: "-"` // Pages: cat.Pages["fr"][0] is the first French page.
}

// NewCategory returns an empty category with Pages initialized
func NewCategory(siteinfo Siteinfo) *Category {
	cat := &Category{}
	cat.Pages = make(map[string][]*Page)
	for locale := range siteinfo.LocalePaths {
		cat.Pages[locale] = make([]*Page, 0)
	}
	return cat
}

// mdTree returns the tree of all pages in markdown format
func (cat *Category) mdTree(prefix string, showPages bool, locale, localePath string) []byte {
	str := fmt.Sprintf("%s* [%s >](%s)\n", prefix, cat.Name, path.Clean(path.Join(localePath, cat.Path(), "index.html")))
	for _, subCat := range cat.SubCategories {
		str += string(subCat.mdTree("\t"+prefix, showPages, locale, localePath))
	}
	if showPages {
		for _, page := range cat.Pages[locale] {
			if page.Basename != "index" {
				str += fmt.Sprintf("%s\t* [%s](%s)\n", prefix, page.Title, path.Clean(path.Join(localePath, page.Path())))
			}
		}
	}
	return []byte(str)
}

// NavHelper returns the tree returned by mdTree, converted to Html format.
func (cat Category) NavHelper(page *Page, showPages bool, locale, localePath string) string {
	return string(Html(cat.mdTree("", showPages, locale, localePath), page, localePath))
}

// FindParent returns the parent category a given file should go in.
// A nil error and a nil parent mean the given path is the root.
func (tree *Category) FindParent(fpath string) (*Category, error) {
	if fpath == ":root:" {
		return nil, nil
	}

	fpath = path.Dir(fpath)

	if fpath == "/" {
		return tree, nil
	}

	pathElems := strings.Split(fpath, "/")[1:]
	parent := tree
	for progress := true; progress && len(pathElems) > 0; {
		progress = false
		for _, subCat := range parent.SubCategories {
			if subCat.Basename == pathElems[0] {
				parent = subCat
				pathElems = pathElems[1:]
				progress = true
				break
			}
		}
	}
	if len(pathElems) > 0 {
		return nil, fmt.Errorf("unable to find suitable parent")
	} else {
		return parent, nil
	}
}

// FilterByTags returns all pages, of a category and its subcategories recursively,
// that match at least one of a given set of tags.
func (cat *Category) FilterByTags(tags []string, locale string) (pages []*Page) {
	for _, page := range cat.Pages[locale] {
		for _, pageTag := range page.Tags {
			for _, testTag := range tags {
				if pageTag == testTag {
					pages = append(pages, page)
					break
				}
			}
		}
	}
	for _, subCat := range cat.SubCategories {
		pages = append(pages, subCat.FilterByTags(tags, locale)...)
	}
	return
}

// FilterByTag wraps FilterByTags for just one tag.
func (cat *Category) FilterByTag(tag string, locale string) []*Page {
	return cat.FilterByTags([]string{tag}, locale)
}

// PageCount returns the total number of pages included in a category and its subcategories.
func (cat *Category) PageCount(locale string) int {
	count := len(cat.Pages[locale])
	for _, subCat := range cat.SubCategories {
		count += subCat.PageCount(locale)
	}
	return count
}

// CategoryCount returns the total number of subcategories included in a category and its subcategories.
func (cat *Category) CategoryCount() int {
	count := len(cat.SubCategories)
	for _, subCat := range cat.SubCategories {
		count += subCat.CategoryCount()
	}
	return count
}

// Path returns a slash-seperated path for the category, starting from the root.
func (cat *Category) Path() string {
	if cat == nil || cat.Parent == nil || cat.Basename == "/" {
		return "/"
	}
	return cat.Parent.Path() + cat.Basename + "/"
}

// Tags returns all the tags present in pages in the category and all subcategories.
func (cat *Category) Tags(locale string) []string {
	tagsMap := make(map[string]bool)
	for _, page := range cat.Pages[locale] {
		for _, tag := range page.Tags {
			tagsMap[tag] = true
		}
	}
	for _, subCat := range cat.SubCategories {
		for _, subTag := range subCat.Tags(locale) {
			tagsMap[subTag] = true
		}
	}
	var tags []string
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return tags
}

// RecentPages returns a list of n pages maximum from the category, most recent first.
func (cat *Category) RecentPages(n int, locale string) (pages []*Page) {
	for catQueue := []*Category{cat}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		pages = append(pages, catQueue[0].Pages[locale]...)
	}
	sort.Slice(pages, func(i, j int) bool {
		ti, err := time.Parse("2006-01-02", pages[i].Date)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return false
		}
		tj, err := time.Parse("2006-01-02", pages[j].Date)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return true
		}
		return ti.After(tj)
	})
	if n >= 0 && len(pages) > n {
		return pages[:n]
	}
	return pages
}
