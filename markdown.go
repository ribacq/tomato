// Tomato static website generator
// Copyright Quentin Ribac, 2018
// Free software license can be found in the LICENSE file.

package main

import (
	"io"

	"gopkg.in/russross/blackfriday.v2"
)

const (
	usedExtensions = blackfriday.Tables | blackfriday.FencedCode | blackfriday.Footnotes | blackfriday.Autolink
)

// Html wraps the blackfriday markdown converter. It takes and returns a slice of bytes.
func Html(content []byte, page *Page, localePath string) []byte {
	renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
		AbsolutePrefix: page.PathToRoot(localePath),
		Flags:          blackfriday.FootnoteReturnLinks,
		FootnoteReturnLinkContents: "<sup>&uarr;</sup>",
	})
	return blackfriday.Run(content, blackfriday.WithRenderer(renderer), blackfriday.WithExtensions(usedExtensions))
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
