package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func QueryWebsite(url string, success chan<- time.Duration, failure chan<- bool) {
	start := time.Now()
	_, err := http.Get(url)
	if err != nil {
		failure <- true
		return
	}
	//defer resp.Body.Close()
	success <- time.Since(start)
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Usage: go run main.go <url> <number of requests> <timeout>")
		return
	}

	url := os.Args[1]

	numRequests, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	timeout, err := strconv.Atoi(os.Args[3])
	if err != nil {
		panic(err)
	}

	success := make(chan time.Duration, numRequests)
	failure := make(chan bool, numRequests)

	start := time.Now()

	for i := 0; i < numRequests; i++ {
		go QueryWebsite(url, success, failure)
	}

	var total time.Duration
	var min = time.Duration(timeout) * time.Second
	var max time.Duration
	fails := 0

	for i := 0; i < numRequests; i++ {
		select {
		case s := <-success:
			if min > s {
				min = s
			}
			if max < s {
				max = s
			}
			total += s
		case <-failure:
			fails++
		case <-time.After(time.Duration(timeout) * time.Second):
			fmt.Println("Timeout")
			fails++
		}
	}

	fmt.Printf("Longest Query: %v\n", max)
	fmt.Printf("Shortest Query: %v\n", min)
	fmt.Printf("Average query: %v\n", total/(time.Duration(numRequests-fails)))
	fmt.Printf("Unsuccessful requests: %d\n", fails)
	fmt.Printf("Total time elapsed: %v\n", time.Since(start))
}
