# Tomato: static blog generator

[![Build Status](https://travis-ci.org/ribacq/tomato.svg?branch=master)](https://travis-ci.org/ribacq/tomato)
[![Coverage Status](https://coveralls.io/repos/github/ribacq/tomato/badge.svg?branch=master)](https://coveralls.io/github/ribacq/tomato?branch=master)
[![GoDoc](https://godoc.org/github.com/ribacq/tomato?status.svg)](https://godoc.org/github.com/ribacq/tomato)

This is a simple static blog generator written in Go. Content will be written in Markdown. Json is used for the site and categories data.

## External ressources
* Markdown to html conversion: [Blackfriday V2](https://github.com/russross/blackfriday/tree/v2.0.0)
* I18n: [qor I18n](https://github.com/qor/i18n)
* Site colors: [Material Palette](https://materialpalette.com/)

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

## Input structure
* site/
	* siteinfo.json
	* pages/
		* catinfo.json
		* index.md
		* page1.md
		* cat1/
			* catinfo.json
			* foo.english-basename.en.md
			* foo.basename-francais.fr.md
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

## Site files
### siteinfo.json
This is the file that defines the site’s main properties. It is written in JSON and looks like this:

```json
{
	"locales": {
		"en": {
			"path": "/",
			"title": "Tomato",
			"subtitle": "I am a subtitle.",
			"description": "Lorem ipsum dolor sit amet...",
			"copyright": "© 2018, Quentin Ribac, public domain"
		},
		"fr": {
			...
		}
	},
	"authors": [
		{
			"name": "Quentin Ribac",
			"email": "my.email@provider.com"
		},
		{
			...
		}
	]
}
```

### catinfo.json
In order for a category to be indexed by tomato, the corresponding directory must contain a `catinfo.json`. It can define the following fields:

```json
{
	"basename": "cat",
	"name": "Category name",
	"description": "I am a category.",
	"unlisted": false
}
```

`basename` is how the category will appear in the URL. If it is not specified, the name of the directory where the `catinfo.json` is will be used.

`unlisted`, if set to `true`, means that the category will exist but hidden in the website, and will not appear in menus. If not specified it is set to `false`.

### Pages
The content of your site will be written in pages (or articles) which are Markdown files in a category directory. They have to be named something like: `foo.basename.en.md`, where:

* `foo` is an ID linking all versions of your pages in different languages if the basename is different. You can omit this ID if the basename stays the same, as in `index.en.md` and `index.fr.md`;
* `basename` will appear in the URL, the converted file will be `basename.html`;
* `en` here stands for the English language, this has to be a locale defined in `siteinfo.json`;
* `.md` indicates the file is in Markdown, this is compulsory, as only `.md` files are detected by Tomato.

As for the content of the file, it must begin with the following lines:

```markdown
#!author: Alice, Bob
#!date: 2018-09-05
#!tags: cats, memes
#!draft

# Page title goes here
Lorem ipsum dolor sit amet...
```

* `#!author: Alice, Bob` can indicate a comma seperated list of authors, or a single one. The author names have to be exactly those defined in `siteinfo.json;
* `#!date: 2018-09-05` has to indicate a date in `YYYY-MM-DD` format. It will be the displayed and sorting date of the article and is here so that you can make changes in the file later without them causing the page to go on top of the list on the website’s home page;
* `#!tags: foo, bar` can contain any strings, comma-seperated;
* `#!draft` is optional and means that the page will be ignored by Tomato and will not appear in the website at all;
* Do not forget the space between `#` and the title of the page after the meta-data, otherwise it will not be detected.

## Internationalization (i18n)
### siteinfo.json
The first file to change in defining locales is `siteinfo.json`:


```json
{
	"locales": {
		"en": {
			"path": "/",
			...
		},
		"fr": {
			"path": "/fr",
			...
		}
	}
	...
}
```

This means that the English version of the website will be at the root: `mysite.com/` and the French version under `mysite.com/fr`.

### catinfo.json
Category files can be written in two different ways: with one version that will be applied to all locales, or with one version per locale:

```json
{
	"name": "blog",
	...
}
```

or

```json
{
	"en": { "name": "01/January", ... },
	"fr": { "name": "01/Janvier", ... }
}
```

### Templates
Locale files must be defined for the **templates**, in YAML format. [example/templates/locales/](example/templates/locales) provides locale files for the example templates in English and French. They look like:

```yaml
en:
    locale_name: English
    full_page:
        header:
            page: Page
            all_tags: All tags
```

### Links
Internal links **must** use the locale path prefixes defined in `siteinfo.json`. This means you have to write `[my link](/fr/page.html)` instead of just `[my link](/page.html)` to stay on the French version, if you have defined the French locale path to `/fr`. This is so because links to images and media will still be like `![alt text](/media/img/plop.png)` without locale prefix, whatever the current locale is, and it also allows for cross-language links.

## Internal process
* Initialize empty tree
* Read /siteinfo.json
* Read category files (all catinfo.json)
	* Add to tree as a Category struct
* Read page files (\*.md)
	* Detect locale and ID
	* Parse meta and content
	* Add to tree as a Page struct
* For each locale:
	* Create Page structs for categories that lack an index
	* Create Page structs for all tags
* Load templates
* For each locale:
	* Generate html pages
* Copy /media
* Copy /assets
