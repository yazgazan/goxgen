package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/yazgazan/goxgen/src/example"
	"github.com/yazgazan/goxgen/src/example/views"
)

func main() {
	listen := ":8973"

	flag.StringVar(&listen, "listen", listen, "")
	flag.Parse()

	srv := &http.Server{
		Addr: listen,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html; charset=utf-8")
			w.WriteHeader(200)

			ctx := example.ExampleContext()
			ctx.GetText = example.TranslateFunc(ctx.Page.Lang)

			res := views.Index(ctx)

			n, err := res.WriteTo(w)
			if err != nil && n == 0 {
				log.Printf("Error: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if err != nil {
				log.Printf("Error: %v", err)
				return
			}
		}),
	}

	log.Printf("listening on %s\n", listen)
	log.Fatal(srv.ListenAndServe())
}
