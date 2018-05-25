package main

import (
	"io"

	"gopkg.in/russross/blackfriday.v2"
)

var (
	usedExtensions = blackfriday.Tables | blackfriday.FencedCode | blackfriday.Footnotes | blackfriday.Autolink
)

// Html wraps the blackfriday markdown converter. It takes and returns a slice of bytes.
func Html(content []byte, page Page) []byte {
	return blackfriday.Run(content, blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{AbsolutePrefix: page.PathToRoot()})), blackfriday.WithExtensions(usedExtensions))
}

type literalRenderer struct{}

func (r literalRenderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if entering && node.Parent != nil && node.Parent.Type != blackfriday.Heading {
		w.Write(node.Literal)
	}
	if !entering && node.Type == blackfriday.Paragraph {
		w.Write([]byte(" "))
	}
	return blackfriday.GoToNext
}

func (r literalRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {}
func (r literalRenderer) RenderFooter(w io.Writer, ast *blackfriday.Node) {}

// Raw strips all markdown formatting from the content.
func Raw(content []byte) []byte {
	renderer := literalRenderer{}
	return blackfriday.Run(content, blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(usedExtensions))
}
