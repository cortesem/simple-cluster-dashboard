package main

import (
	"log"
	"net/http"
	"simple-cluster-dashboard/pkg/dashboard"
)

func main() {
	dashboard := dashboard.HTMLPage{}
	dashboard.New("k3s.local")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Handling connection from: ", r.Host)
		dashboard.Generate(w)
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "favicon.ico")
	})

	log.Println("Listening on port 8008...")

	http.ListenAndServe(":8008", nil)
}
