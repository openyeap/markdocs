package main

import (
	"strings"
	"testing"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func TestInit(t *testing.T) {
	site := NewSite("site.json")
	site.Public()
	site.Server()
}
func TestMarkdown(t *testing.T) {
	md := []byte(`# header1
## header2	
!note test

!important $2^2$

![x](http://byeap.com)
	`)
	// create markdown parser with extensions
	extensions := parser.Attributes | parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock | parser.Titleblock
	p := parser.NewWithExtensions(extensions)

	p.RegisterInline('!', func(p *parser.Parser, data []byte, offset int) (int, ast.Node) {
		str := strings.Split(strings.ToLower(string(data[1:])), " ")
		switch str[0] {
		case "note":
			link := &ast.HTMLBlock{
				Leaf: ast.Leaf{

					Literal: []byte(`
<div class="admonition note">
	<p class="admonition-title">Note</p>
	<p>` + str[1] + `</p>
</div>`),
				},
			}
			return len(data), link
		case "important":
			link := &ast.HTMLBlock{
				Leaf: ast.Leaf{

					Literal: []byte(`
<div class="admonition important">
	<p class="admonition-title">Important</p>
	<p>` + str[1] + `</p>
</div>`),
				},
			}
			return len(data), link
		default:
			return 0, nil
		}

	})

	doc := p.Parse(md)
	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CommonFlags
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	print(string(markdown.Render(doc, renderer)))
}
