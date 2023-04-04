package main

import (
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", http.FileServer(http.Dir("D:\\go\\src")))
	if err != nil {
		panic(err)
	}
}
