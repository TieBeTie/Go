//go:build !solution

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	urls := os.Args[1:]

	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(string(body))
	}
}
