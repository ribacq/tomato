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
		fmt.Println("Error: please specify an input directory.")
		return
	}

	// exit if incorrect input directory was given
	if DirectoryExists(os.Args[1]) {
		inputDir = path.Clean(os.Args[1])
	} else {
		fmt.Println("Error: " + os.Args[1] + " is not a directory")
		return
	}
	if inputDir == "/" {
		fmt.Println("Error: cannot use /.")
		return
	}

	fmt.Println("Using " + inputDir)

	// set and maybe create output directory
	outputDir = inputDir + "_html"
	if FileExists(outputDir) {
		fmt.Println("Error: " + outputDir + " already exists and is not a directory.")
		return
	}
	if DirectoryExists(outputDir) {
		fmt.Println("Deleting pre-existing output directory")
		rmOutput := exec.Command("rm", "-rf", outputDir)
		err := rmOutput.Start()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = rmOutput.Wait()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	err := os.Mkdir(outputDir, 0775)
	if err != nil {
		fmt.Println(err)
		return
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
			fmt.Println("Error: could not open /siteinfo.json")
			return
		}
		jsonDecoder := json.NewDecoder(f)
		err = jsonDecoder.Decode(&siteinfo)
		if err != nil {
			fmt.Println("Error: incorrect /siteinfo.json")
			return
		}
	} else {
		fmt.Println("Error: no /siteinfo.json found")
		return
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
		fmt.Println(err)
		return
	}
	fmt.Printf("%v categories found\n", tree.CategoryCount())

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
			titleRE := regexp.MustCompile("# .+\\n")
			authorRE := regexp.MustCompile("#!author: .+\\n")
			dateRE := regexp.MustCompile("#!date: (\\d{4}-\\d{2}-\\d{2})\\n")
			tagsRE := regexp.MustCompile("#!tags: .+\\n")
			draftRE := regexp.MustCompile("#!draft\\n")
			featuredImageLinkRE := regexp.MustCompile("!!\\[(.+)\\]\\((.+)\\)")

			title := strings.Trim(strings.TrimPrefix(string(titleRE.Find(content)), "#"), " \n")
			author := strings.Trim(strings.TrimPrefix(string(authorRE.Find(content)), "#!author:"), " \n")
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
			au, err := siteinfo.FindAuthor(author)
			if err != nil {
				return err
			}
			page := &Page{nil, strings.TrimSuffix(path.Base(fpath), ".md"), title, au, date, tags, draft, content, pathToFeaturedImage}

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
		fmt.Println(err)
		return
	}
	fmt.Printf("%v pages found\n", tree.PageCount())

	// load templates
	fullPageTemplate := template.Must(template.ParseFiles(inputDir + "/templates/full_page.html"))
	pageListTemplate := template.Must(template.ParseFiles(inputDir + "/templates/page_list.html"))

	// walk tree
	fmt.Println("\n\x1b[1mGenerating html...\x1b[0m")
	for catQueue := []*Category{tree}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		// create subdirectories
		for _, subCat := range catQueue[0].SubCategories {
			if !DirectoryExists(outputDir + subCat.Path()) {
				err := os.Mkdir(outputDir+subCat.Path(), 0755)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}

		// create page files
		for _, page := range catQueue[0].Pages {
			pageFile, err := os.OpenFile(outputDir+page.Path(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
			if err != nil {
				fmt.Println(err)
				return
			}

			arg := map[string]interface{}{
				"Siteinfo": siteinfo,
				"Page":     page,
				"Tree":     tree,
			}
			err = fullPageTemplate.ExecuteTemplate(pageFile, "Header", arg)
			if err != nil {
				fmt.Println(err)
				return
			}
			contentTemplate := template.Must(template.New("Content").Parse(string(page.ContentHelper())))
			err = contentTemplate.ExecuteTemplate(pageFile, "Content", arg)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = fullPageTemplate.ExecuteTemplate(pageFile, "Footer", arg)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println(page.Path())
		}
	}

	// make categories index.html files with catinfo.json if there is no index.html yet
	fmt.Println("\n\x1b[1mGenerating index.html files for categories lacking them...\x1b[0m")
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

		catFile, err := os.OpenFile(outputDir+catQueue[0].Path()+"index.html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Println(err)
			return
		}

		arg := map[string]interface{}{
			"Siteinfo": siteinfo,
			"Page": &Page{
				Title:    "Category : " + catQueue[0].Name,
				Author:   &siteinfo.Authors[0],
				Tags:     catQueue[0].Tags(),
				Category: catQueue[0],
				Basename: "index",
			},
			"Tree": tree,
		}
		err = fullPageTemplate.ExecuteTemplate(catFile, "Header", arg)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = pageListTemplate.ExecuteTemplate(catFile, "PageList", map[string]interface{}{
			"Pages": catQueue[0].Pages,
			"Page": &Page{
				Category: catQueue[0],
				Basename: "index",
			},
			"Category": catQueue[0],
			"Title":    "Category: " + catQueue[0].Name,
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		err = fullPageTemplate.ExecuteTemplate(catFile, "Footer", arg)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(catQueue[0].Path() + "index.html")
	}

	// make tags html files
	if !DirectoryExists(outputDir + "/tag") {
		err := os.Mkdir(outputDir+"/tag", 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	for _, tag := range tree.Tags() {
		tagFile, err := os.OpenFile(outputDir+"/tag/"+tag+".html", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Println(err)
			return
		}
		arg := map[string]interface{}{
			"Siteinfo": siteinfo,
			"Page": &Page{
				Title:    "Tag: " + tag,
				Author:   &siteinfo.Authors[0],
				Tags:     []string{tag},
				Category: tree,
				Basename: "tag/" + tag,
			},
			"Tree": tree,
		}
		err = fullPageTemplate.ExecuteTemplate(tagFile, "Header", arg)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = pageListTemplate.ExecuteTemplate(tagFile, "PageList", map[string]interface{}{
			"Pages": tree.FilterByTags([]string{tag}),
			"Title": "Tag: " + tag,
			"Page": &Page{
				Category: tree,
				Basename: "tag/" + tag,
			},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		err = fullPageTemplate.ExecuteTemplate(tagFile, "Footer", arg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("\n\x1b[1mCopying resource directories...\x1b[0m")
	// copy /media
	if DirectoryExists(inputDir + "/media") {
		fmt.Println("Copying /media")
		cpMedia := exec.Command("cp", "-Rv", inputDir+"/media", outputDir)
		err = cpMedia.Start()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = cpMedia.Wait()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// copy /assets
	if DirectoryExists(inputDir + "/assets") {
		fmt.Println("Copying /assets")
		cpAssets := exec.Command("cp", "-Rv", inputDir+"/assets", outputDir)
		err = cpAssets.Start()
		if err != nil {
			fmt.Println(err)
			return
		}
		err = cpAssets.Wait()
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
