package main

import (
	"fmt"
	"log"

	clangformat "github.com/javorszky/go-diff-clang/pkg/clang-format"
)

func main() {

	format, lc, err := clangformat.IdealClangFormatFile()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("the ideal clang format file changing %d lines"+
		" is this:\n\n%s\n", lc, format)
}
