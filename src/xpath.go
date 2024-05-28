package main

import (
	"os"
	"path/filepath"
	"strings"
)

type X struct {
	Site
	Page
}

func FileExtesion(path string) (string, int) {
	for i := len(path) - 1; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			return path[i:], i
		}
	}
	return "", -1
}
func Rel(site Site, elem string) string {

	rel, err := filepath.Rel(site.SrcDir, elem)
	if err != nil {
		return strings.ReplaceAll(elem, "\\", "/")
	}
	return strings.ReplaceAll(rel, "\\", "/")
}
func (x X) AbsRel(elem string) string {

	root := filepath.Join(x.Site.SrcDir, x.Page.Path)
	if x.Page.Type != "index" {
		root = filepath.Dir(root)
	}
	target := filepath.Join(x.Site.SrcDir, elem)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return elem
	}
	return strings.ReplaceAll(rel, "\\", "/")
}
func (x X) Cwd() string {
	root := filepath.Join(x.Site.Root(), x.Site.SrcDir)
	target := filepath.Join(x.Site.Root(), x.Page.Path)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return strings.ReplaceAll(x.Page.Path, "\\", "/")
	}
	return strings.ReplaceAll(rel, "\\", "/")
}
