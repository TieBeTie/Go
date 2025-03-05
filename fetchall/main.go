//go:build !solution

package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

func main() {
	urls := os.Args[1:]
	ch := make(chan string)
	max_seconds := 0.0

	for _, url := range urls {
		go func(url string) {
			start := time.Now()
			resp, err := http.Get(url)
			if err != nil {
				log.Printf("Get %s: %v", url, err)
				ch <- fmt.Sprintf("error: %s", url)
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("Read %s: %v", url, err)
				ch <- fmt.Sprintf("error: %s", url)
				return
			}
			ch <- fmt.Sprintf("%0.2fs\t%d  %s", time.Since(start).Seconds(), len(body), url)
			max_seconds = math.Max(max_seconds, time.Since(start).Seconds())
		}(url)
	}
	for range urls {
		fmt.Printf("%s\n", <-ch)
	}
	fmt.Printf("%0.2fs elapsed\n", max_seconds)
}
