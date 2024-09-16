package main

import (
	"fmt"
	"net/http"
)

func main() {

	http.Handle("/user/", NewOtherApi())

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)
}
