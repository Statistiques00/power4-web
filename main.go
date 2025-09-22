package main

import (
	"fmt"
	"html/template"
	"net/http"
)

var tmpl = template.Must(template.ParseFiles("index.html"))

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, nil)
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
	fmt.Print("Le server tourne")
}
