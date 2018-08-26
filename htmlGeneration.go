// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"text/template"
)

// GenerateIndividualPages creates HTML files and calls the templates for each page defined in the website
func GenerateIndividualPages(siteinfo Siteinfo, tree *Category, templates *template.Template, inputDir, outputDir, locale string) error {
	for catQueue := []*Category{tree}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		// create subdirectories
		for _, subCat := range catQueue[0].SubCategories {
			if !DirectoryExists(path.Join(outputDir, siteinfo.LocalePaths[locale], subCat.Path())) {
				err := os.Mkdir(path.Join(outputDir, siteinfo.LocalePaths[locale], subCat.Path()), 0755)
				if err != nil {
					return err
				}
			}
		}

		// create page files
		for _, page := range catQueue[0].Pages {
			pageFile, err := os.OpenFile(path.Join(outputDir, siteinfo.LocalePaths[locale], page.Path()), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
			if err != nil {
				return err
			}

			arg := map[string]interface{}{
				"Siteinfo": siteinfo,
				"Locale":   locale,
				"Page":     page,
				"Tree":     tree,
			}
			err = templates.ExecuteTemplate(pageFile, "Header", arg)
			if err != nil {
				return err
			}
			templates = template.Must(templates.Parse("{{ define \"Content\" }}" + page.ContentHelper(siteinfo.LocalePaths[locale]) + "{{ end }}"))
			err = templates.ExecuteTemplate(pageFile, "Content", arg)
			if err != nil {
				return err
			}
			template.Must(templates.Parse("{{ define \"Content\" }}{{ end }}"))
			err = templates.ExecuteTemplate(pageFile, "Footer", arg)
			if err != nil {
				return err
			}

			//fmt.Println(page.Path())
		}
	}

	return nil
}

// GenerateCategoryPages creates index.html files for categories lacking them
func GenerateCategoryPages(siteinfo Siteinfo, tree *Category, templates *template.Template, inputDir, outputDir, locale string) error {
	for catQueue := []*Category{tree}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		// if there is already an index.md: do nothing
		mustContinue := false
		for _, page := range catQueue[0].Pages {
			if page.Basename == "index" {
				mustContinue = true
				break
			}
		}
		if mustContinue {
			continue
		}

		// create file
		catFile, err := os.OpenFile(path.Join(outputDir, siteinfo.LocalePaths[locale], catQueue[0].Path(), "index.html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			return err
		}

		// create argument for templates
		arg := map[string]interface{}{
			"Siteinfo": siteinfo,
			"Locale":   locale,
			"Page": &Page{
				Category: catQueue[0],
				Basename: "index",
				Title:    "Category : " + catQueue[0].Name,
				Authors:  []*Author{&siteinfo.Authors[0]},
				Tags:     catQueue[0].Tags(),
			},
			"Tree": tree,
		}

		// execute templates
		err = templates.ExecuteTemplate(catFile, "Header", arg)
		if err != nil {
			return err
		}
		err = templates.ExecuteTemplate(catFile, "PageList", arg)
		if err != nil {
			return err
		}
		err = templates.ExecuteTemplate(catFile, "Footer", arg)
		if err != nil {
			return err
		}

		//fmt.Println(catQueue[0].Path() + "index.html")
	}

	return nil
}

// GenerateTagPages create a ‘tag’ directory and tag pages in it.
func GenerateTagPages(siteinfo Siteinfo, tree *Category, templates *template.Template, inputDir, outputDir, locale string) error {
	// create tag directory
	tagDir := path.Join(outputDir, siteinfo.LocalePaths[locale], "tag")
	if !DirectoryExists(tagDir) {
		err := os.Mkdir(tagDir, 0755)
		if err != nil {
			return err
		}
	}
	for _, tag := range tree.Tags() {
		// create tag file
		tagFile, err := os.OpenFile(path.Join(tagDir, tag+".html"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			return nil
		}

		// create template argument
		arg := map[string]interface{}{
			"Siteinfo": siteinfo,
			"Locale":   locale,
			"Page": &Page{
				Category: &Category{
					Parent:      tree,
					Basename:    "tag",
					Name:        "Tags",
					Description: "Pages tagged with " + tag,
					Pages:       tree.FilterByTags([]string{tag}),
				},
				Basename: tag,
				Title:    "Tag: " + tag,
				Authors:  []*Author{&siteinfo.Authors[0]},
				Tags:     []string{tag},
			},
			"Tree": tree,
		}

		// execute templates
		err = templates.ExecuteTemplate(tagFile, "Header", arg)
		if err != nil {
			return err
		}
		err = templates.ExecuteTemplate(tagFile, "PageList", arg)
		if err != nil {
			return err
		}
		err = templates.ExecuteTemplate(tagFile, "Footer", arg)
		if err != nil {
			return err
		}
	}

	return nil
}
