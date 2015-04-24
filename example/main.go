package main

import (
	"html/template"
	"net/http"

	"github.com/shelakel/view"
)

var Templates = view.New(50, map[string]view.View{
	"pages.home": template.Must(template.New("layout.html").ParseFiles(
		// register templates from least to most specific e.g. layout, nested layout, page
		"./views/shared/layout.html", // the layout must be first
		"./views/home/home.html")),
})

type PageWrapper struct{ Title string }

func ServePage(viewName, pageTitle string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := Templates.Render(viewName, w, &PageWrapper{Title: pageTitle}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func main() {
	http.Handle("/", ServePage("pages.home", "Home"))
	http.ListenAndServe(":8080", nil)
}
