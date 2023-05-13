package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("New connection")
		fmt.Fprintf(w, "If you see this response... success!")
	})

	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":8080", nil)
}
