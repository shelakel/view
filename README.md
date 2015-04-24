# view: Simple Go template manager

Feel free to vendor the source without attribution.

# Usage

See example/main.go

### main.go

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

### layout.html

	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>{{ .Title }}</title>
	</head>
	<body>
	{{ template "body" . }}
	</body>
	</html>
	{{ define "body" }}{{ end }}

### home.html

	{{ define "body" }}
	<h1>{{ .Title }}</h1>
	{{ end }}

Tips:
 * Note the {{ define "body" }} on both *layout.html* and *home.html*. Having {{ define "body" }} on *layout.html*  makes the {{ define "body" }} on *home.html* optional.
 * Use (*Manager).MustRegister to dynamically add or replace templates - for example, you could design a sub system that would reparse the templates when a dependant file changed or create a singleton package to host the template manager and register package specific views in func init() {}.

# License
BSD 3-clause
