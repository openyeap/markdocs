package main

import (
	"testing"
)

func TestInit(t *testing.T) {
	site := NewSite("site.json")
	site.Public()
	site.Server()
}
