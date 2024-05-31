package main

import (
	"net/url"
	"path/filepath"
)

type X struct {
	Site
	Page
}

func FileExtesion(path string) (string, int) {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:], i
		}
	}
	return "", -1
}
func Rel(site Site, elem string) string {

	rel, err := filepath.Rel(site.SrcDir, elem)
	if err != nil {
		return filepath.ToSlash(elem)
	}
	return filepath.ToSlash(rel)
}
func (x X) AbsRel(elem string) string {
	root := filepath.Dir(filepath.Join(x.Site.SrcDir, x.Page.Path))
	target := filepath.Join(x.Site.SrcDir, elem)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return elem
	}
	return filepath.ToSlash(rel)
}
func (x X) Cwd() string {
	root := filepath.Join(x.Site.Root(), x.Site.SrcDir)
	target := filepath.Join(x.Site.Root(), x.Page.Path)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return filepath.ToSlash(x.Page.Path)
	}
	return filepath.ToSlash(rel)
}
func (x X) Encode(rawURL string) string {

	return url.PathEscape(rawURL)

}
