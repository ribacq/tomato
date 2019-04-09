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
// Name and Description are fetched from a `catinfo.json` file that should exist in of every directory.
// Basename is the bit that goes in the URL.
type Category struct {
	Parent        *Category                      `json: "-"`
	SubCategories []*Category                    `json: "-"`
	Realname      string                         `json: "-"`
	Locales       map[string]*CategoryLocaleData `json: "locales"`
}

// CategoryLocaleData holds data of a category that changes with the locale
type CategoryLocaleData struct {
	Basename    string  `json: "basename"`
	Name        string  `json: "name"`
	Description string  `json: "description"`
	Unlisted    bool    `json: "unlisted"`
	Pages       []*Page `json: "-"`
}

// NewCategory returns an empty category with Locales initialized
func NewCategory(siteinfo Siteinfo) *Category {
	cat := &Category{}
	cat.Locales = make(map[string]*CategoryLocaleData)
	for locale := range siteinfo.Locales {
		cat.Locales[locale] = &CategoryLocaleData{}
	}
	return cat
}

// Tree returns the closest parent category whose parent is nil.
func (cat *Category) Tree() *Category {
	ret := cat
	for ret.Parent != nil {
		ret = ret.Parent
	}
	return ret
}

// TagCategory returns the category with basename "tag" at the root.
func (cat *Category) TagCategory() (*Category, error) {
	tree := cat.Tree()
	for _, subCat := range tree.SubCategories {
		for locale := range subCat.Locales {
			if subCat.Locales[locale].Basename == "tag" {
				return subCat, nil
			}
		}
	}
	return nil, fmt.Errorf("Unable to find ‘tag’ category at root of tree.")
}

// IsUnder returns whether a category is or is a descendent of a given category.
func (cat *Category) IsUnder(testCat *Category) bool {
	for curCat := cat; curCat != nil; curCat = curCat.Parent {
		if curCat == testCat {
			return true
		}
	}
	return false
}

// IsEmpty returns whether a category is fully empty in a given locale
func (cat *Category) IsEmpty(locale string) bool {
	return len(cat.Locales[locale].Pages) == 0 && len(cat.SubCategories) == 0 && cat.PageCount(locale) == 0
}

// mdTree returns the tree of all pages in markdown format
func (cat *Category) mdTree(prefix string, showPages bool, locale, localePath string) []byte {
	str := fmt.Sprintf("%s* [%s >](%s)\n", prefix, cat.Locales[locale].Name, path.Clean(path.Join(localePath, cat.Path(locale), "index.html")))
	for _, subCat := range cat.SubCategories {
		if !subCat.Locales[locale].Unlisted && !subCat.IsEmpty(locale) {
			str += string(subCat.mdTree("\t"+prefix, showPages, locale, localePath))
		}
	}
	if showPages {
		for _, page := range SortByRecent(cat.Locales[locale].Pages) {
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
			if subCat.Realname == pathElems[0] {
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
	for _, page := range cat.Locales[locale].Pages {
		if page.Unlisted || page.Category != cat {
			continue
		}
	tagFor:
		for _, pageTag := range page.Tags {
			for _, testTag := range tags {
				if pageTag == testTag {
					pages = append(pages, page)
					break tagFor
				}
			}
		}
	}
	for _, subCat := range cat.SubCategories {
		if !subCat.Locales[locale].Unlisted {
			pages = append(pages, subCat.FilterByTags(tags, locale)...)
		}
	}
	return
}

// FilterByTag wraps FilterByTags for just one tag.
func (cat *Category) FilterByTag(tag string, locale string) []*Page {
	return cat.FilterByTags([]string{tag}, locale)
}

// PageCount returns the total number of pages included in a category and its subcategories.
func (cat *Category) PageCount(locale string) int {
	count := 0
	for _, page := range cat.Locales[locale].Pages {
		if !page.Unlisted && page.Category == cat {
			count++
		}
	}
	for _, subCat := range cat.SubCategories {
		count += subCat.PageCount(locale)
	}
	return count
}

// CategoryCount returns the total number of subcategories included in a category and its subcategories.
func (cat *Category) CategoryCount(locale string) int {
	count := 0
	for _, subCat := range cat.SubCategories {
		if !subCat.Locales[locale].Unlisted {
			count += 1 + subCat.CategoryCount(locale)
		}
	}
	return count
}

// Path returns a slash-seperated path for the category, starting from the root.
func (cat *Category) Path(locale string) string {
	if cat == nil || cat.Parent == nil || cat.Locales[locale].Basename == "/" {
		return "/"
	}
	return cat.Parent.Path(locale) + cat.Locales[locale].Basename + "/"
}

// Tags returns all the tags present in pages in the category and all subcategories.
func (cat *Category) Tags(locale string) []string {
	tagsMap := make(map[string]bool)
	for _, page := range cat.Locales[locale].Pages {
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
		if !cat.Locales[locale].Unlisted && catQueue[0].Locales[locale].Unlisted {
			continue
		}
		for _, page := range catQueue[0].Locales[locale].Pages {
			if !page.Unlisted {
				alreadyThere := false
				for _, page2 := range pages {
					if page2 == page {
						alreadyThere = true
						break
					}
				}
				if !alreadyThere {
					pages = append(pages, page)
				}
			}
		}
	}
	pages = SortByRecent(pages)
	if n >= 0 && len(pages) > n {
		return pages[:n]
	}
	return pages
}

// SortByRecent returns a copy of the slice sorted by recent first.
func SortByRecent(pages []*Page) (ret []*Page) {
	ret = append(ret, pages...)
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].Basename == "index" || ret[j].Basename == "index" {
			return ret[i].Basename != "index"
		}
		ti, err := time.Parse("2006-01-02", ret[i].Date)
		if err != nil {
			fmt.Fprintln(os.Stderr, ret[i].Path(), err)
			return false
		}
		tj, err := time.Parse("2006-01-02", ret[j].Date)
		if err != nil {
			fmt.Fprintln(os.Stderr, ret[i].Path(), err)
			return true
		}
		return ti.After(tj)
	})
	return
}
