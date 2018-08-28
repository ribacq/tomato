#!author: Quentin Ribac
#!date: 2018-05-20
#!tags: blog

# À propos
Cette page est à propos de ce blog. Elle en décrit les auteurs, etc. Les auteurs sont :

{{ range .Siteinfo.Authors }}
	{{ .Helper }}
{{ end }}

[&rarr; markdown](/markdown.html)

{{ template "Disqus" }}
