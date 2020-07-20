package main

import (
	"log"
	"os"

	exampletext "github.com/yazgazan/goxgen/src/example-text"

	"github.com/yazgazan/goxgen/src/example-text/views"
)

func main() {
	ctx := exampletext.ExampleContext()
	ctx.GetText = exampletext.TranslateFunc(ctx.Page.Lang)

	res := views.Index(ctx)

	_, err := res.WriteTo(os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
