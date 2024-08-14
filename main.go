package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Hello, World!")
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		tmpl := template.Must(template.ParseFiles("test.html"))
		tmpl.Execute(w, nil)
	}
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		ID := r.FormValue("ID")
		MSG := r.FormValue("MSG")
		htmlstr := fmt.Sprintf("<div><h1>ID: %s</h1><h1>MSG: %s</h1></div>", ID, MSG)
		tmpl, _ := template.New("h").Parse(htmlstr)
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/", helloHandler)
	fmt.Println("server running on 'http://localhost:8080'")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
