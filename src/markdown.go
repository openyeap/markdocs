package main

import (
	"fmt"

	"github.com/gomarkdown/markdown"
)

func main4() {

	md := []byte("## markdown document")
	output := markdown.ToHTML(md, nil, nil)
	fmt.Println(output)
}
