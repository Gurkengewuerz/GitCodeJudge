package markdown

import (
	"bytes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
)

var MD = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
	goldmark.WithRendererOptions(
		html.WithHardWraps(),
		html.WithXHTML(),
	),
)

func FormatMarkdownToHTML(md string) (template.HTML, error) {
	var htmlBuf bytes.Buffer
	if err := MD.Convert([]byte(md), &htmlBuf); err != nil {
		return "", err
	}
	return template.HTML(htmlBuf.String()), nil
}
