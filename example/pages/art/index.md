#!author: Jirsad
#!date: 2018-05-17
#!tags: art

# Art
This category lists all of my personal art projects.

<ul>
{{ $pathToRoot := .Page.PathToRoot }}
{{ range .Page.Category.Pages }}
	{{ if ne .Basename "index" }}
		<li><a href="{{ $pathToRoot }}{{ .Path }}">{{ .Title }}</a></li>
	{{ end }}
{{ end }}
</ul>
