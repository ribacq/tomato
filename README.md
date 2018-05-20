# static blog generator

This is a simple static blog generator written in Go.

## TODO
* most recent pages in menu
* translation of template text

## Process
* Initialize empty tree
* Read /siteinfo.json
* Read category files (all catinfo.json)
	* Add to tree as a Category struct
* Read page files (\*.md)
	* Parse meta and content
	* Add to tree as a Page struct
* Make html header, menu and footer
* Generate html pages
* Generate category pages for those that have no `index.md`
* Generate tags html files
* Copy /media
* Copy /assets

## External ressources
* Markdown to html conversion: [Blackfriday V2](https://github.com/russross/blackfriday/tree/v2.0.0)
* Site colors: [Material Palette](https://materialpalette.com/)

## Input structure
```
site/
  siteinfo.json
  pages/
	catinfo.json
    index.md
	cat1/
	  catinfo.json
	  page1.md
	  page2.md
    en/
	  catinfo.json
	  index.md
	  cat1/
	    catinfo.json
		page1.md
		page2.md
  media/
    img/
	  img1.png
    doc/
	  cv.pdf
	data/
	  plop.tar.gz
  assets/
    header_template.html
	menu_template.html
	footer_template.html
	style.css
	main.js
```

## Output structure
```
site/
  index.html
  cat1/
	index.html
    page1.html
	page2.html
  en/
    index.html
	cat1/
	  index.html
	  page1.html
	  page2.html
  tag/
    tag1.html
	tag2.html
  media/
	img/
	  img1.png
	doc/
	  cv.pdf
	data/
	  plop.tar.gz
```
