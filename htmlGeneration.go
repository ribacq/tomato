// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"text/template"

	"github.com/qor/i18n"
)

// GenerateIndividualPages creates HTML files and calls the templates for each page defined in the website
func GenerateIndividualPages(siteinfo *Siteinfo, tree *Category, templates *template.Template, inputDir, outputDir string, locales *i18n.I18n, locale string) (n int, err error) {
	for catQueue := []*Category{tree}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		// skip empty category
		if catQueue[0].IsEmpty(locale) {
			continue
		}

		// create subdirectories
		for _, subCat := range catQueue[0].SubCategories {
			if !DirectoryExists(path.Join(outputDir, siteinfo.Locales[locale].Path, subCat.Path(locale))) {
				err = os.Mkdir(path.Join(outputDir, siteinfo.Locales[locale].Path, subCat.Path(locale)), 0755)
				if err != nil {
					return n, err
				}
			}
		}

		// create page files
		for _, page := range catQueue[0].Locales[locale].Pages {
			// skip page if its category is not the one it’s accessed by
			if catQueue[0] != page.Category {
				continue
			}

			// create file
			pageFile, err := os.OpenFile(path.Join(outputDir, siteinfo.Locales[locale].Path, page.Path()), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
			if err != nil {
				return n, err
			}

			// prepare template argument
			arg := map[string]interface{}{
				"Siteinfo": *siteinfo,
				"Locale":   locale,
				"Page":     page,
				"Tree":     tree,
			}

			// header template
			err = templates.ExecuteTemplate(pageFile, "Header", arg)
			if err != nil {
				return n, err
			}

			// content template
			templates = template.Must(templates.Parse("{{ define \"Content\" }}{{ $localePath := (index .Siteinfo.Locales .Locale).Path }}" + page.ContentHelper(siteinfo.Locales[locale].Path) + "{{ end }}"))
			err = templates.ExecuteTemplate(pageFile, "Content", arg)
			if err != nil {
				return n, err
			}
			template.Must(templates.Parse("{{ define \"Content\" }}{{ end }}"))

			// footer template
			err = templates.ExecuteTemplate(pageFile, "Footer", arg)
			if err != nil {
				return n, err
			}

			// generated page count
			n++
		}
	}

	return n, nil
}
