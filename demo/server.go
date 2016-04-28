package main

import (
	"fmt"
	"github.com/justgolang/gracego"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	})

	err := gracego.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
