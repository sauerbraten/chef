package main

import (
	"embed"
	"io/fs"
)

var (
	//go:embed templates css
	embedded embed.FS

	templates fs.FS
	css       fs.FS
)

func init() {
	var err error

	templates, err = fs.Sub(embedded, "templates")
	if err != nil {
		panic(err)
	}

	css, err = fs.Sub(embedded, "css")
	if err != nil {
		panic(err)
	}
}
