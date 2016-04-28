package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8989")
	checkError(err)

	i := 1
	for {
		_, err = conn.Write([]byte("test-content-" + strconv.FormatInt(int64(i), 10)))
		i++
		checkError(err)

		response := make([]byte, 256)
		readLength, err := conn.Read(response)
		checkError(err)

		if readLength > 0 {
			fmt.Println("[client] server response:", string(response))
			time.Sleep(1 * time.Second)
		}
	}
}

func checkError(err error) {
	if err != nil {
		log.Fatal("fatal error: " + err.Error())
	}
}
