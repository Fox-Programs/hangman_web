package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {

	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/test", indexhandler)
	http.HandleFunc("/pendu", penduhandler)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	musique := http.FileServer(http.Dir("./musique"))
	http.Handle("/musique/", http.StripPrefix("/musique/", musique))

	fmt.Printf("test")
	if err := http.ListenAndServe(":7080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexhandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/test" {
		http.Error(w, "existe pas déso", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Oulah tu veux me hacker ptite pute", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "pâtes")

}

func penduhandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/pendu"{
		http.Error(w, "nn déso", http.StatusNotFound)
		return
	}

	if r.Method != "GET"{
		http.Error(w, "non", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, "pendu.html")
	
}