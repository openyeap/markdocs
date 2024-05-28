package main

import (
	"fmt"
	"sort"
	"testing"
)

func TestABC(t *testing.T) {
	pages := []string{"分布式存储", "sss", "2020-06-08-文档构建"}
	less := func(i, j int) bool {

		return pages[j] > pages[i]
	}

	sort.Slice(pages, less)

	fmt.Println(pages)
	fmt.Println([]byte(pages[0]))
	fmt.Println(Sha1("hello"))
	fmt.Println(Sha2("hello"))
}
