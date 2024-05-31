package main

import (
	"fmt"
	"testing"
)

func TestABC(t *testing.T) {

	fmt.Println(Md5("hello"))
	fmt.Println(Sha1("hello"))
	fmt.Println(Sha2("hello"))
}
