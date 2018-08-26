#!author: Quentin Ribac
#!date: 2018-05-20
#!tags: blog

# About
This page is about this blog. It describes its authors and stuff. The authors are:

[&rarr; markdown](/en/markdown.html)

{{ range .Siteinfo.Authors }}
	{{ .Helper }}
{{ end }}

{{ template "Disqus" }}
