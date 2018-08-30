# Tomato: static blog generator

[![Build Status](https://travis-ci.org/ribacq/tomato.svg?branch=master)](https://travis-ci.org/ribacq/tomato)
[![Coverage Status](https://coveralls.io/repos/github/ribacq/tomato/badge.svg?branch=master)](https://coveralls.io/github/ribacq/tomato?branch=master)
[![GoDoc](https://godoc.org/github.com/ribacq/tomato?status.svg)](https://godoc.org/github.com/ribacq/tomato)

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
* I18n: [Qor I18n](https://github.com/qor/i18n)
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

## Internationalization (i18n)
Locales must first be defined in `siteinfo.json`:

```json
{
	...
	"localePaths": {
		"en": "/",
		"fr": "/fr"
	},
	...
}
```

This means that the English version of the website will be at the root: `mysite.com/` and the French version under `mysite.com/fr/`.

Locale files must be defined for the templates, in YAML format. [example/templates/locales/](example/templates/locales) provides locale files for the example templates in English and French. They look like:

```yaml
en:
    locale_name: English
    full_page:
        header:
            page: Page
            all_tags: All tags
```

Then in order to translate the pages themselves, you have to create one file per locale: `pages/cat1/page.en.md` and `pages/cat1/page.fr.md` if you defined English and French. `pages/cat1/page.md` will be set to the locale defined with `"/"` in `localePaths` in `siteinfo.json`.

Internal links **must** use the locale path prefixes defined in `siteinfo.json`. This means you have to write `[my link](/fr/page.html)` instead of just `[my link](/page.html)` to stay on the French version, if you have defined the French locale path to `/fr`. This is so because links to images and media will still be like `![alt text](/media/img/plop.png)` without locale prefix, whatever the current locale is, and it also allows for cross-language links.

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
