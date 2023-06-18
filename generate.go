package main

import (
	"net/http"
	"simple-cluster-dashboard/pkg/dashboard"
)

func main() {
	dashboard := dashboard.HTMLPage{}
	dashboard.New("k3s.local")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dashboard.Generate(w)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	http.ListenAndServe(":8008", nil)
}
