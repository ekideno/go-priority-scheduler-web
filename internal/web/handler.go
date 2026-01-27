package web

import (
	"html/template"
	"net/http"
)

func RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderIndex(w)
	})

}

func renderIndex(w http.ResponseWriter) {
	tmpl := template.Must(template.ParseFiles("internal/web/templates/index.html"))

	tmpl.Execute(w, nil)
}

func RegisterAPIRoutes(mux *http.ServeMux) {
	apiMux := http.NewServeMux()

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}
