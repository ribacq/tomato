# Tomato: static blog generator

[![Build Status](https://travis-ci.org/ribacq/tomato.svg?branch=master)](https://travis-ci.org/ribacq/tomato)
[![Coverage Status](https://coveralls.io/repos/github/ribacq/tomato/badge.svg?branch=master)](https://coveralls.io/github/ribacq/tomato?branch=master)
[![GoDoc](https://godoc.org/github.com/ribacq/tomato?status.svg)](https://godoc.org/github.com/ribacq/tomato)
x
This is a simple static blog generator written in Go. Content will be written in Markdown. Json is used for the site and categories data.

## How-to
```bash
git clone https://github.com/ribacq/tomato
cd tomato
name="my_website_name"
mv example "$name"
# edit "$name"’s content
go build tomato.go
./tomato "$name"
firefox "${name}_html/index.html"
```
## External ressources
* Markdown to html conversion: [Blackfriday V2](https://github.com/russross/blackfriday/tree/v2.0.0)
* Site colors: [Material Palette](https://materialpalette.com/)

## Input structure
* site/
	* siteinfo.json
	* pages/
		* catinfo.json
		* index.md
		* page1.md
		* cat1/
			* catinfo.json
			* page2.md
	* media/
		* img/
			* cat.png
		* doc/
			* cv.pdf
		* data/
			* archive.tar.gz
	* assets/
		* style.css
		* main.js
	* templates/
		* full_page.html
		* page_list.html

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
