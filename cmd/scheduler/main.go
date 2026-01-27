package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	s := &http.Server{
		Addr: ":8080",
		// Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Println("Server working")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world")
	})
	log.Fatal(s.ListenAndServe())

}
