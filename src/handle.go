package main

import (
	"net/http"
)

func (site *Site) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.FileServer(http.Dir(site.DstDir)).ServeHTTP(w, r)
}
