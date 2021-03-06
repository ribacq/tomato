// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

/*
Tomato is a static website generator.
*/
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"text/template"
)

// main is the entry point for the program.
// This program should be called with one console-line argument: the path to the input directory.
func main() {
	// set input and output directories
	fmt.Println("\x1b[1mSetting input and output directories...\x1b[0m")
	var inputDir, outputDir string

	// exit if no input directory was given
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: please specify an input directory.")
		os.Exit(1)
	}

	// exit if incorrect input directory was given
	if DirectoryExists(os.Args[1]) {
		inputDir = path.Clean(os.Args[1])
	} else {
		fmt.Fprintln(os.Stderr, "Error: "+os.Args[1]+" is not a directory.")
		os.Exit(1)
	}
	if inputDir == "/" {
		fmt.Fprintln(os.Stderr, "Error: cannot use root (/) as input directory.")
		os.Exit(1)
	}

	fmt.Println("Input: " + inputDir)

	// set and maybe create output directory
	if len(os.Args) > 2 && os.Args[2] != inputDir {
		outputDir = path.Clean(os.Args[2])
	} else {
		outputDir = inputDir + "_html"
	}
	if FileExists(outputDir) {
		fmt.Fprintln(os.Stderr, "Error: "+outputDir+" already exists and is not a directory.")
		os.Exit(1)
	}

	fmt.Println("Output: " + outputDir)

	if DirectoryExists(outputDir) {
		fmt.Println("Deleting pre-existing output directory")
		rmOutput := exec.Command("rm", "-rf", outputDir)
		err := rmOutput.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = rmOutput.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	err := os.Mkdir(outputDir, 0775)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// load /siteinfo.json
	fmt.Println("\n\x1b[1mLoading /siteinfo.json...\x1b[0m")
	var siteinfo Siteinfo
	if FileExists(inputDir + "/siteinfo.json") {
		f, err := os.Open(inputDir + "/siteinfo.json")
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: could not open /siteinfo.json")
			os.Exit(1)
		}
		jsonDecoder := json.NewDecoder(f)
		err = jsonDecoder.Decode(&siteinfo)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: incorrect /siteinfo.json")
			os.Exit(1)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Error: no /siteinfo.json found")
		os.Exit(1)
	}
	// create directories from .LocalePaths
	for locale := range siteinfo.Locales {
		if !DirectoryExists(path.Join(outputDir, siteinfo.Locales[locale].Path)) {
			err := os.Mkdir(path.Join(outputDir, siteinfo.Locales[locale].Path), 0775)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
	fmt.Printf("Done, %v locales, %v authors found.\n", len(siteinfo.Locales), len(siteinfo.Authors))

	// load template locales
	locales := LoadLocales(inputDir + "/templates/locales")

	// initialize empty tree
	tree := NewCategory(siteinfo)

	// read category files (all catinfo.json)
	fmt.Println("\n\x1b[1mLoading categories...\x1b[0m")
	err = WalkDir(inputDir, func(fpath string) error {
		if path.Base(fpath) == "catinfo.json" {
			// open file
			// fmt.Println(fpath)
			f, err := os.Open(fpath)
			if err != nil {
				return err
			}

			// create category struct
			cat := NewCategory(siteinfo)

			// find path and Basename
			fpath = path.Dir(strings.TrimPrefix(fpath, inputDir+"/pages"))
			basename := path.Base(fpath)
			cat.Realname = basename
			if fpath == "/" {
				fpath = ":root:"
			}

			// json decoder
			jsonDecoder := json.NewDecoder(f)

			// try json => CategoryLocaleData
			var locale string
			for l := range siteinfo.Locales {
				locale = l
				break
			}
			err = jsonDecoder.Decode(cat.Locales[locale])
			if err == nil && len(cat.Locales[locale].Name) > 0 {
				// put this value in all locales
				for l := range siteinfo.Locales {
					if l != locale {
						*cat.Locales[l] = *cat.Locales[locale]
					}
				}
			} else {
				// try json => Category
				f.Seek(0, 0)
				err = jsonDecoder.Decode(&cat.Locales)
				if err != nil {
					return err
				}
			}

			// set basename where not set yet
			for locale := range siteinfo.Locales {
				if cat.Locales[locale].Basename == "" {
					cat.Locales[locale].Basename = basename
				}
			}

			// locate parent
			parent, err := tree.FindParent(fpath)
			if err != nil {
				return err
			}

			if parent == nil {
				// parent is nil: this is the root category
				for locale := range siteinfo.Locales {
					*tree.Locales[locale] = *cat.Locales[locale]
				}
			} else {
				// parent is not nil: insert category
				parent.SubCategories = append(parent.SubCategories, cat)
				cat.Parent = parent
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for locale := range siteinfo.Locales {
		fmt.Printf("%v: %v categories found\n", locale, 1+tree.CategoryCount(locale))
	}

	// read page files (*.md)
	fmt.Println("\n\x1b[1mLoading pages...\x1b[0m")
	err = WalkDir(inputDir, func(fpath string) error {
		if strings.HasSuffix(path.Base(fpath), ".md") {
			// detect locale, fallback to root locale defined in siteinfo.json
			var locale string
			for localeCandidate := range siteinfo.Locales {
				if siteinfo.Locales[localeCandidate].Path == "/" {
					locale = localeCandidate
				}
				if strings.HasSuffix(path.Base(fpath), "."+localeCandidate+".md") {
					locale = localeCandidate
					break
				}
			}
			if locale == "" {
				return fmt.Errorf("Unable to detect locale for %v", fpath)
			}

			basename := strings.TrimSuffix(strings.TrimSuffix(path.Base(fpath), ".md"), "."+locale)
			basenameParts := strings.Split(basename, ".")
			id := basename
			if len(basenameParts) > 1 {
				id = basenameParts[0]
				basename = strings.TrimPrefix(basename, id+".")
			}

			// load file content
			content, err := ReadFile(fpath)
			if err != nil {
				return err
			}

			// parse meta and remove them from content
			titleRE := regexp.MustCompile("(?m)^# .+$")
			authorRE := regexp.MustCompile("(?m)^#!author: (.+)(, .+)*$")
			shortSummaryRE := regexp.MustCompile("(?m)^#!short-summary: .+$")
			dateRE := regexp.MustCompile("(?m)^#!date: (\\d{4}-\\d{2}-\\d{2})$")
			tagsRE := regexp.MustCompile("(?m)^#!tags: .+$")
			draftRE := regexp.MustCompile("(?m)^#!draft$")
			featuredImageLinkRE := regexp.MustCompile("!!\\[(.+)\\]\\((.+)\\)")

			title := strings.Trim(strings.TrimPrefix(string(titleRE.Find(content)), "#"), " \n")
			shortSummary := strings.Trim(strings.TrimPrefix(string(shortSummaryRE.Find(content)), "#!short-summary:"), " \n")
			authorsNames := strings.Split(strings.Trim(strings.TrimPrefix(string(authorRE.Find(content)), "#!author:"), " \n"), ", ")
			date := strings.Trim(strings.TrimPrefix(string(dateRE.Find(content)), "#!date:"), " \n")
			tags := strings.Split(strings.Trim(strings.TrimPrefix(string(tagsRE.Find(content)), "#!tags:"), " \n"), ",")
			for i := 0; i < len(tags); i++ {
				tags[i] = strings.Trim(tags[i], " \n")
				if tags[i] == "" {
					tags = append(tags[:i], tags[i+1:]...)
					i--
				}
			}

			draft := strings.Trim(strings.TrimPrefix(string(draftRE.Find(content)), "#!"), " \n") == "draft"
			if draft {
				fmt.Printf("Skipping draft: ‘%s’\n", fpath)
				return nil
			}

			pathToFeaturedImage := ""
			submatches := featuredImageLinkRE.FindSubmatch(content)
			if len(submatches) >= 2 {
				pathToFeaturedImage = string(submatches[2])
			}

			content = authorRE.ReplaceAll(content, []byte{})
			content = shortSummaryRE.ReplaceAll(content, []byte{})
			content = dateRE.ReplaceAll(content, []byte{})
			content = tagsRE.ReplaceAll(content, []byte{})
			content = draftRE.ReplaceAll(content, []byte{})
			content = featuredImageLinkRE.ReplaceAll(content, []byte("![$1]($2)"))

			// add to tree as a Page struct
			var authors []*Author
			for i := range authorsNames {
				author, err := siteinfo.FindAuthor(authorsNames[i])
				if err != nil {
					println(fpath)
					return err
				}
				authors = append(authors, author)
			}
			page := &Page{
				ID:                  id,
				Basename:            basename,
				Title:               title,
				ShortSummary:        shortSummary,
				Authors:             authors,
				Date:                date,
				Tags:                tags,
				Draft:               draft,
				Content:             content,
				PathToFeaturedImage: pathToFeaturedImage,
				Locale:              locale,
			}

			parent, err := tree.FindParent(strings.TrimPrefix(fpath, inputDir+"/pages"))
			if err != nil {
				return err
			}
			if parent == nil {
				tree.Locales[locale].Pages = append(tree.Locales[locale].Pages, page)
				page.Category = tree
			} else {
				parent.Locales[locale].Pages = append(parent.Locales[locale].Pages, page)
				page.Category = parent
			}

			/*/ special title for site home page
			if page.Category == tree && page.Basename == "index" {
				page.Title = siteinfo.Locales[locale].Subtitle
			}*/
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for locale := range siteinfo.Locales {
		fmt.Printf("%v: %v pages found\n", locale, tree.PageCount(locale))
	}

	// create categories for tags
	tagCat := NewCategory(siteinfo)
	tagCat.Parent = tree
	tree.SubCategories = append(tree.SubCategories, tagCat)
	for locale := range siteinfo.Locales {
		tagCat.Locales[locale].Basename = "tag"
		tagCat.Locales[locale].Name = "Tags"
		tagCat.Locales[locale].Unlisted = true
	}
	for locale := range siteinfo.Locales {
	toNextTag:
		for _, tag := range tree.Tags(locale) {
			// skip if it was already created
			for _, cat := range tagCat.SubCategories {
				if cat.Locales[locale].Basename == tag {
					continue toNextTag
				}
			}
			cat := NewCategory(siteinfo)
			cat.Parent = tagCat
			for locale2 := range siteinfo.Locales {
				cat.Locales[locale2].Basename = tag
				cat.Locales[locale2].Name = tag
				cat.Locales[locale2].Unlisted = true
				cat.Locales[locale2].Pages = tree.FilterByTag(tag, locale2)
			}
			tagCat.SubCategories = append(tagCat.SubCategories, cat)
		}
	}

	// for each locale, make index pages for categories lacking them
	for locale := range siteinfo.Locales {
		for catQueue := []*Category{tree}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
			// if there is already an index.md: do nothing
			mustContinue := false
			for _, page := range catQueue[0].Locales[locale].Pages {
				if page.Basename == "index" && page.Category == catQueue[0] {
					mustContinue = true
					break
				}
			}
			if mustContinue {
				continue
			}

			// skip empty categories
			if catQueue[0].PageCount(locale) == 0 && !catQueue[0].IsUnder(tagCat) {
				continue
			}

			// create category page
			catPage := NewCategoryPage(catQueue[0], &siteinfo, locales, locale)

			// change title for tag pages
			if catPage.Category.IsUnder(tagCat) {
				catPage.Title = string(locales.T(locale, "tags.page_list_name", catPage.Category.Locales[locale].Name))
			}

			// add the page to its category
			catQueue[0].Locales[locale].Pages = append(catQueue[0].Locales[locale].Pages, catPage)
		}
	}

	// load templates
	templates := template.New("tomatoTemplates")
	templates.Funcs(map[string]interface{}{
		"t": func(locale, key string, args ...interface{}) string {
			return string(locales.T(locale, key, args...))
		},
		"join": func(paths ...string) string {
			return path.Clean(path.Join(paths...))
		},
	})
	_, err = templates.ParseGlob(inputDir + "/templates/*.html")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error when parsing templates: ", err)
		os.Exit(1)
	}

	// generate the html pages for all locales
	for locale := range siteinfo.Locales {
		fmt.Println("\n\x1b[1mLocale: " + locale + ", in " + siteinfo.Locales[locale].Path + "\x1b[0m")
		n, err := GenerateIndividualPages(&siteinfo, tree, templates, inputDir, outputDir, locales, locale)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		fmt.Printf("%v html files generated\n", n)
	}

	fmt.Println("\n\x1b[1mCopying resource directories...\x1b[0m")
	// copy /media
	if DirectoryExists(inputDir + "/media") {
		fmt.Println("Copying /media")
		cpMedia := exec.Command("cp", "-Rv", inputDir+"/media", outputDir)
		err = cpMedia.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = cpMedia.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// copy /assets
	if DirectoryExists(inputDir + "/assets") {
		fmt.Println("Copying /assets")
		cpAssets := exec.Command("cp", "-Rv", inputDir+"/assets", outputDir)
		err = cpAssets.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		err = cpAssets.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
