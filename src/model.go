package main

import (
	"os"
	"text/template"
	"time"
)

type Item struct {
	Title    string `json:"title"`
	Path     string `json:"path"`
	Children []Item `json:"children,omitempty"`
}

type Setting struct {
	Meta    map[string]interface{}
	Title   string
	Layout  string
	Url     string
	Date    time.Time
	DstDir  string
	SrcDir  string
	Exclude string
	Toc     []Item
}

func (setting *Setting) Root() string {
	if _, yes := setting.Meta["root"]; !yes {
		value, _ := os.Getwd()
		setting.Meta["root"] = value
	}
	return setting.Meta["root"].(string)
}

type Site struct {
	Setting
	Template *template.Template
	Manifest []Page
}

// <item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml"/>
type Page struct {
	Path     string
	Title    string
	Date     time.Time
	Layout   string
	Content  string
	Type     string
	Meta     map[string]interface{}
	Children []Page
}

func (page *Page) Sort() string {
	if value, has := page.Meta["sort"]; has {
		v, y := value.(string)
		if y {
			if page.Type == "index" {
				return "朱" + v
			}
			return v

		}
	}
	if page.Type == "index" {
		return "朱" + page.Title
	}
	return page.Title
}
