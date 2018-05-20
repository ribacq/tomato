#!author: Quentin Ribac
#!date: 2018-05-20
#!tags: blog, meta

# About
This page is about this blog. It describes its authors and stuff. The authors are:

{{ range .Siteinfo.Authors }}
	{{ .Helper }}
{{ end }}
