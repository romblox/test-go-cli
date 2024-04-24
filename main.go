package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func queryUrl(urlStr *string, numThreads *int, timeout *int, queryTimes chan<- time.Duration, errorTimes chan<- bool) {
	for i := 0; i < *numThreads; i++ {
		go func() {
			startTime := time.Now()

			client := &http.Client{
				Timeout: time.Duration(*timeout) * time.Second,
			}

			_, err := client.Get(*urlStr)
			queryTime := time.Since(startTime)

			if err != nil {
				fmt.Printf("Error: %s", err)
				errorTimes <- true
				return
			}

			queryTimes <- queryTime
		}()
	}
}

func getCliArgs() (urlStr *string, numThreads *int, timeout *int, err error) {
	urlStr = flag.String("url", "", "URL to query")
	numThreads = flag.Int("threads", 1, "Number of concurrent threads")
	timeout = flag.Int("timeout", 0, "Timeout in seconds (optional)")
	flag.Parse()

	parsedURL, err := url.Parse(*urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		err = fmt.Errorf("provide a valid url")
		return
	}

	if *numThreads <= 0 {
		err = fmt.Errorf("number of concurrent threads must be greater than 0")
		return
	}

	if *timeout < 0 {
		err = fmt.Errorf("timeout must be equal or greater than 0")
		return
	}
	return
}

func main() {
	var (
		longestQuery     time.Duration
		shortestQuery    time.Duration
		totalQueryTime   time.Duration
		successfulReqs   int
		unsuccessfulReqs int
	)

	urlStr, numThreads, timeout, err := getCliArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	queryTimes := make(chan time.Duration, *numThreads)
	errorTimes := make(chan bool, *numThreads)

	queryUrl(urlStr, numThreads, timeout, queryTimes, errorTimes)

	for i := 0; i < *numThreads; i++ {
		select {
		case queryTime := <-queryTimes:
			if queryTime > longestQuery {
				longestQuery = queryTime
			}
			if queryTime < shortestQuery || shortestQuery == 0 {
				shortestQuery = queryTime
			}
			totalQueryTime += queryTime
			successfulReqs++
		case <-errorTimes:
			unsuccessfulReqs++
		}
	}

	averageQueryTime := totalQueryTime / time.Duration(successfulReqs)

	fmt.Printf("Longest query time: %v\n", longestQuery)
	fmt.Printf("Shortest query time: %v\n", shortestQuery)
	fmt.Printf("Average query time: %v\n", averageQueryTime)
	fmt.Printf("Number of successful requests: %d\n", successfulReqs)
	fmt.Printf("Number of unsuccessful requests: %d\n", unsuccessfulReqs)
}
