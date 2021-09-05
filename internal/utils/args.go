package utils

import (
	"flag"
	"os"
)

func ParsePathArg() (string, []os.DirEntry) {
	p := "."

	flag.Parse()
	if flag.NArg() == 1 {
		p = flag.Arg(0)
	}

	files, err := os.ReadDir(p)
	if err != nil {
		panic(err)
	}

	return p, files
}
