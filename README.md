Package ghttp provides a graceful shutdown and restart for Go.

Usage
=============
Just use `ghttp.ListenAndServe` instead of `http.ListenAndServe`.  
Send `SIGUSR2` signal to a go process that is using ghttp in order to graceful restart and  send `SIGTERM` in order to graceful shutdonw.
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


Installation
=============
`go get github.com/justgolang/ghttp`  
use `go get -u` to update the package.  