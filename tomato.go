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

	fmt.Println("Using " + inputDir)

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

	fmt.Println("Output will be written to " + outputDir)

	// initialize empty tree
	tree := &Category{}

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
	for locale := range siteinfo.LocalePaths {
		if !DirectoryExists(path.Join(outputDir, siteinfo.LocalePaths[locale])) {
			err := os.Mkdir(path.Join(outputDir, siteinfo.LocalePaths[locale]), 0775)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}
	fmt.Printf("Done, %v authors found.\n", len(siteinfo.Authors))

	// read category files (all catinfo.json)
	fmt.Println("\n\x1b[1mLoading categories...\x1b[0m")
	err = WalkDir(inputDir, func(fpath string) error {
		if path.Base(fpath) == "catinfo.json" {
			fmt.Println(fpath)
			f, err := os.Open(fpath)
			if err != nil {
				return err
			}
			jsonDecoder := json.NewDecoder(f)
			var cat Category
			err = jsonDecoder.Decode(&cat)
			if err != nil {
				return err
			}
			fpath = path.Dir(strings.TrimPrefix(fpath, inputDir+"/pages"))
			cat.Basename = path.Base(fpath)
			if fpath == "/" {
				fpath = ":root:"
			}
			parent, err := tree.FindParent(fpath)
			if err != nil {
				return err
			}
			if parent == nil {
				tree.Name = cat.Name
				tree.Description = cat.Description
				tree.Basename = cat.Basename
			} else {
				parent.SubCategories = append(parent.SubCategories, &cat)
				cat.Parent = parent
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%v categories found\n", 1+tree.CategoryCount())

	// read page files (*.md)
	fmt.Println("\n\x1b[1mLoading pages...\x1b[0m")
	err = WalkDir(inputDir, func(fpath string) error {
		if strings.HasSuffix(path.Base(fpath), ".md") {
			// load file content
			content, err := ReadFile(fpath)
			if err != nil {
				return err
			}

			// parse meta and remove them from content
			titleRE := regexp.MustCompile("(?m)^# .+$")
			authorRE := regexp.MustCompile("(?m)^#!author: (.+)(, .+)*$")
			dateRE := regexp.MustCompile("(?m)^#!date: (\\d{4}-\\d{2}-\\d{2})$")
			tagsRE := regexp.MustCompile("(?m)^#!tags: .+$")
			draftRE := regexp.MustCompile("(?m)^#!draft$")
			featuredImageLinkRE := regexp.MustCompile("!!\\[(.+)\\]\\((.+)\\)")

			title := strings.Trim(strings.TrimPrefix(string(titleRE.Find(content)), "#"), " \n")
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
			pathToFeaturedImage := ""
			submatches := featuredImageLinkRE.FindSubmatch(content)
			if len(submatches) >= 2 {
				pathToFeaturedImage = string(submatches[2])
			}

			content = authorRE.ReplaceAll(content, []byte{})
			content = dateRE.ReplaceAll(content, []byte{})
			content = tagsRE.ReplaceAll(content, []byte{})
			content = draftRE.ReplaceAll(content, []byte{})
			content = featuredImageLinkRE.ReplaceAll(content, []byte("![$1]($2)"))

			// add to tree as a Page struct
			var authors []*Author
			for i := range authorsNames {
				author, err := siteinfo.FindAuthor(authorsNames[i])
				if err != nil {
					return err
				}
				authors = append(authors, author)
			}
			page := &Page{nil, strings.TrimSuffix(path.Base(fpath), ".md"), title, authors, date, tags, draft, content, pathToFeaturedImage}

			if page.Draft {
				fmt.Printf("Skipping draft: ‘%s’\n", page.Title)
				return nil
			}

			parent, err := tree.FindParent(strings.TrimPrefix(fpath, inputDir+"/pages"))
			if err != nil {
				return err
			}
			if parent == nil {
				tree.Pages = append(tree.Pages, page)
				page.Category = tree
			} else {
				parent.Pages = append(parent.Pages, page)
				page.Category = parent
			}

			fmt.Println(fpath)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%v pages found\n", tree.PageCount())

	// load locales
	locales := LoadLocales(inputDir + "/templates/locales")

	// load templates
	templates := template.New("tomatoTemplates")
	templates.Funcs(map[string]interface{}{
		"t": func(locale, key string, args ...interface{}) string {
			return string(locales.T(locale, key, args...))
		},
	})
	_, err = templates.ParseGlob(inputDir + "/templates/*.html")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error when parsing templates: ", err)
		os.Exit(1)
	}

	// for each locale
	for locale := range siteinfo.LocalePaths {
		fmt.Println("\n\x1b[1mLocale: " + locale + ", in " + siteinfo.LocalePaths[locale] + "\x1b[0m")
		// generate individual html pages
		fmt.Println("Generating html files for individual pages...")
		if err = GenerateIndividualPages(siteinfo, tree, templates, inputDir, outputDir, locale); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		// make categories index.html files with catinfo.json if there is no index.html yet
		fmt.Println("Generating index.html files for categories lacking them...")
		if err = GenerateCategoryPages(siteinfo, tree, templates, inputDir, outputDir, locale); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		// make tags html files
		fmt.Println("Generating html files for tags...")
		if err = GenerateTagPages(siteinfo, tree, templates, inputDir, outputDir, locale); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
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
