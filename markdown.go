package main

import (
	"gopkg.in/russross/blackfriday.v2"
)

// Html wraps the blackfriday markdown converter. It takes and returns a slice of bytes.
func Html(content []byte, page Page) []byte {
	return blackfriday.Run(content, blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{AbsolutePrefix: page.PathToRoot()})), blackfriday.WithExtensions(blackfriday.Tables|blackfriday.FencedCode|blackfriday.Footnotes|blackfriday.Autolink))
}
