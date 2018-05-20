package main

/* main: tomato is a static website generator. */

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"gopkg.in/russross/blackfriday.v2"
)

// Author is the type for an author of the website.
type Author struct {
	Name  string `json: "name"`
	Email string `json: "email"`
}

// Siteinfo contains the site-wide meta. There should be only one of them.
// name and title will be printed in the header,
// description will be printed in the menu,
// copyright will be printed in the footer.
// PathPrefix is a prefix to append to all relative URLs.
// Authors must contain all possible authors for the website.
type Siteinfo struct {
	Name        string   `json: "name"`
	Title       string   `json: "title"`
	Subtitle    string   `json: "subtitle"`
	Description string   `json: "description"`
	Copyright   string   `json: "copyright"`
	Authors     []Author `json: "authors"`
}

// Category represent a category, that is, a directory in the tree.
// Name and description are fetched from a `catinfo.json` file that should
// be located at the root of every directory.
type Category struct {
	Parent        *Category   `json: "-"`
	Basename      string      `json: "-"`
	Name          string      `json: "name"`
	Description   string      `json: "description"`
	SubCategories []*Category `json: "-"`
	Pages         []*Page     `json: "-"`
}

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

// Html wraps the blackfriday markdown converter. It takes and returns a slice of bytes.
func Html(content []byte, page Page) []byte {
	return blackfriday.Run(content, blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{AbsolutePrefix: page.PathToRoot()})), blackfriday.WithExtensions(blackfriday.Tables|blackfriday.FencedCode|blackfriday.Footnotes|blackfriday.Autolink))
}

// Helper prints a html link to an author.
func (author *Author) Helper() string {
	return fmt.Sprintf("<address><a href=\"mailto:%s\">%s</a></address>", author.Email, author.Name)
}

// MainAuthorHelper prints a html link to the first author of the siteinfo.
func (siteinfo Siteinfo) MainAuthorHelper() string {
	return siteinfo.Authors[0].Helper()
}

// CopyrightHelper prints html for the copyright information.
func (siteinfo Siteinfo) CopyrightHelper(page Page) string {
	return string(Html([]byte(siteinfo.Copyright), page))
}

// SubtitleHelper prints html for the site subtitle.
func (siteinfo Siteinfo) SubtitleHelper(page Page) string {
	return string(Html([]byte(siteinfo.Subtitle), page))
}

// DescriptionHelper prints html for the site description.
func (siteinfo Siteinfo) DescriptionHelper(page Page) string {
	return string(Html([]byte(siteinfo.Description), page))
}

// ContentHelper prints the page in html.
func (page Page) ContentHelper() string {
	return string(Html(page.Content, page))
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

// FilterByTags returns all pages, of a category and its subcategories recursively,
// that match at least one of a given set of tags.
func (cat *Category) FilterByTags(tags []string) (pages []*Page) {
	for _, page := range cat.Pages {
		for _, pageTag := range page.Tags {
			for _, testTag := range tags {
				if pageTag == testTag {
					pages = append(pages, page)
				}
			}
		}
	}
	for _, subCat := range cat.SubCategories {
		pages = append(pages, subCat.FilterByTags(tags)...)
	}
	return
}

// PageCount returns the total number of pages included in a category and its subcategories.
func (cat *Category) PageCount() int {
	count := len(cat.Pages)
	for _, subCat := range cat.SubCategories {
		count += subCat.PageCount()
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

// FindAuthor returns an existing author by its name or nil and an error if there is no author with this name.
func (si *Siteinfo) FindAuthor(name string) (*Author, error) {
	for i := range si.Authors {
		if si.Authors[i].Name == name {
			return &si.Authors[i], nil
		}
	}
	return nil, fmt.Errorf("unable to find this author")
}

// Path returns a slash-seperated path for the category, starting from the root.
func (cat *Category) Path() string {
	if cat == nil || cat.Parent == nil {
		return "/"
	}
	return cat.Parent.Path() + cat.Basename + "/"
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

// MDTree returns the tree of all pages in markdown format
func (cat *Category) MDTree(prefix string, showPages bool) []byte {
	str := fmt.Sprintf("%s* [%s >](%sindex.html)\n", prefix, cat.Name, cat.Path())
	for _, subCat := range cat.SubCategories {
		str += string(subCat.MDTree("\t"+prefix, showPages))
	}
	if showPages {
		for _, page := range cat.Pages {
			if page.Basename != "index" {
				str += fmt.Sprintf("%s\t* [%s](%s)\n", prefix, page.Title, page.Path())
			}
		}
	}
	return []byte(str)
}

// NavHelper returns the tree returned by MDTree, converted to Html format.
func (cat Category) NavHelper(page Page, showPages bool) string {
	return string(Html(cat.MDTree("", showPages), page))
}

// Tags returns all the tags present in pages in the category and all subcategories.
func (cat *Category) Tags() []string {
	tagsMap := make(map[string]bool)
	for _, page := range cat.Pages {
		for _, tag := range page.Tags {
			tagsMap[tag] = true
		}
	}
	for _, subCat := range cat.SubCategories {
		for _, subTag := range subCat.Tags() {
			tagsMap[subTag] = true
		}
	}
	var tags []string
	for tag := range tagsMap {
		tags = append(tags, tag)
	}
	return tags
}

// RecentPages returns a list of n pages maximum from the category, most recent first.
func (cat *Category) RecentPages(n int) (pages []*Page) {
	for catQueue := []*Category{cat}; len(catQueue) > 0; catQueue = append(catQueue[1:], catQueue[0].SubCategories...) {
		pages = append(pages, catQueue[0].Pages...)
	}
	sort.Slice(pages, func(i, j int) bool {
		ti, err := time.Parse("2006-01-02", pages[i].Date)
		if err != nil {
			fmt.Println(err)
			return false
		}
		tj, err := time.Parse("2006-01-02", pages[j].Date)
		if err != nil {
			fmt.Println(err)
			return false
		}
		return ti.After(tj)
	})
	if len(pages) > n {
		return pages[:n]
	}
	return pages
}

// FindParent returns the parent category a given file should go in.
// A nil error and a nil parent mean the given path is the root.
func (tree *Category) FindParent(fpath string) (*Category, error) {
	if fpath == "." {
		return nil, nil
	}

	fpath = path.Clean(path.Dir(fpath))

	if fpath == "/" {
		return tree, nil
	}

	pathElems := strings.Split(fpath, "/")[1:]
	parent := tree
	for progress := true; progress && len(pathElems) > 0; {
		progress = false
		for _, subCat := range parent.SubCategories {
			if len(pathElems) == 0 {
				break
			}
			if subCat.Basename == pathElems[0] {
				parent = subCat
				pathElems = pathElems[1:]
				progress = true
			}
		}
	}
	if len(pathElems) > 0 {
		return nil, fmt.Errorf("unable to find suitable parent")
	} else {
		return parent, nil
	}
}

// FileExists returns whether a given name exists and is a regular file.
func FileExists(name string) bool {
	if fi, err := os.Stat(name); err == nil && fi.Mode().IsRegular() {
		return true
	}
	return false
}

// FileExists returns whether a given path exists and is a directory.
func DirectoryExists(name string) bool {
	if fi, err := os.Stat(name); err == nil && fi.Mode().IsDir() {
		return true
	}
	return false
}

// ReadFile reads all the content of a file and returns it as a slice of bytes.
func ReadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	var content []byte
	buff := make([]byte, 1024)
	offset := int64(0)
	for {
		n, err := f.ReadAt(buff, offset)
		offset += int64(n)
		content = append(content, buff[:n]...)
		if err != nil {
			break
		}
	}
	return content, nil
}

// WalkDirFunc is the callback type used by WalkDir.
type WalkDirFunc func(fname string) error

// WalkDir walks a directory tree beginning at the given root.
// In every directory, it first calls the callback on every regular file.
// Then it pushes all subdirectories to the queue.
func WalkDir(root string, callback WalkDirFunc) error {
	var dirQueue []string
	dirQueue = append(dirQueue, root)

	for len(dirQueue) > 0 {
		dir, err := os.Open(dirQueue[0])
		if err != nil {
			fmt.Println(err)
			return err
		}

		names, err := dir.Readdirnames(0)
		if err != nil {
			fmt.Println(err)
			return err
		}

		for _, name := range names {
			if FileExists(dirQueue[0] + "/" + name) {
				err = callback(dirQueue[0] + "/" + name)
				if err != nil {
					fmt.Println(err)
					return err
				}
			} else if DirectoryExists(dirQueue[0] + "/" + name) {
				dirQueue = append(dirQueue, dirQueue[0]+"/"+name)
			}
		}
		dirQueue = dirQueue[1:]
	}
	return nil
}

// main is the entry point for the program.
// This program should be called with one console-line argument: the path to the input directory.
func main() {
	// set input directory
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
	var siteinfo Siteinfo
	if FileExists(inputDir + "/siteinfo.json") {
		fmt.Println("Loading /siteinfo.json")
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

	// read category files (all catinfo.json)
	fmt.Println("Loading categories... ")
	err = WalkDir(inputDir, func(fpath string) error {
		if path.Base(fpath) == "catinfo.json" {
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
			cat.Basename = path.Base(path.Dir(fpath))
			parent, err := tree.FindParent(path.Clean(strings.TrimPrefix(path.Dir(fpath), inputDir+"/pages")))
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
	fmt.Printf("%v categories found\n", 1+tree.CategoryCount())

	// read page files (*.md)
	fmt.Println("Loading pages... ")
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

			content = authorRE.ReplaceAll(content, []byte{})
			content = dateRE.ReplaceAll(content, []byte{})
			content = tagsRE.ReplaceAll(content, []byte{})
			content = draftRE.ReplaceAll(content, []byte{})

			// add to tree as a Page struct
			au, err := siteinfo.FindAuthor(author)
			if err != nil {
				return err
			}
			page := &Page{nil, strings.TrimSuffix(path.Base(fpath), ".md"), title, au, date, tags, draft, content, ""}

			if page.Draft {
				fmt.Printf("Skipping draft: ‘%s’\n", page.Title)
				return nil
			}

			parent, err := tree.FindParent(strings.TrimPrefix(fpath, inputDir+"/"))
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
	fmt.Println("Generating html...")
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
			contentTemplate := template.Must(template.New("Content").Parse(string(Html(page.Content, *page))))
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
		}
	}

	// make categories index.html files with catinfo.json if there is no index.html yet
	fmt.Println("Generating index.html files for categories lacking them...")
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
			"Tags":     catQueue[0].Tags(),
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
			"Tags": []string{tag},
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
