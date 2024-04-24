package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

func main() {
	var (
		longestQuery     time.Duration
		shortestQuery    time.Duration
		totalQueryTime   time.Duration
		successfulReqs   int
		unsuccessfulReqs int
	)

	urlStr := flag.String("url", "", "URL to query")
	numThreads := flag.Int("threads", 1, "Number of concurrent threads")
	timeout := flag.Int("timeout", 0, "Timeout in seconds (optional)")
	flag.Parse()

	// Validate URL format
	parsedURL, err := url.Parse(*urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		fmt.Println("Please provide a valid URL.")
		return
	}

	// Validate number of concurrent threads
	if *numThreads <= 0 {
		fmt.Println("Number of concurrent threads must be > 0")
		return
	}

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a channel to collect query times
	queryTimes := make(chan time.Duration, *numThreads)

	// Start goroutines for concurrent requests
	wg.Add(*numThreads)
	for i := 0; i < *numThreads; i++ {
		go func() {
			defer wg.Done()
			startTime := time.Now()

			// Create an HTTP client with custom timeout
			client := &http.Client{
				Timeout: time.Duration(*timeout) * time.Second,
			}

			// Make an HTTP request
			_, err := client.Get(*urlStr)
			if err != nil {
				fmt.Println("Error:", err)
				unsuccessfulReqs++
				return
			}

			// Calculate query time
			queryTime := time.Since(startTime)
			queryTimes <- queryTime

			// Update statistics
			if queryTime > longestQuery {
				longestQuery = queryTime
			}
			if queryTime < shortestQuery || shortestQuery == 0 {
				shortestQuery = queryTime
			}
			totalQueryTime += queryTime
			successfulReqs++
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(queryTimes)

	// Calculate average query time
	averageQueryTime := totalQueryTime / time.Duration(successfulReqs)

	// Print statistics
	fmt.Printf("Longest query time: %v\n", longestQuery)
	fmt.Printf("Shortest query time: %v\n", shortestQuery)
	fmt.Printf("Average query time: %v\n", averageQueryTime)
	fmt.Printf("Number of successful requests: %d\n", successfulReqs)
	fmt.Printf("Number of unsuccessful requests: %d\n", unsuccessfulReqs)
}
