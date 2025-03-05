//go:build !solution

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	files := os.Args[1:]
	lines := make(map[string]int64)

	for _, file := range files {
		file, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			line := scanner.Text()
			lines[line]++
		}
	}

	for line, count := range lines {
		if count >= 2 {
			fmt.Printf("%d\t%s\n", count, line)
		}
	}
}
