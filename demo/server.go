package main

import (
	"fmt"
	"github.com/justgolang/ghttp"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	})

	err := ghttp.ListenAndServe(":8989", nil)
	if err != nil {
		fmt.Println(err)
	}
}
