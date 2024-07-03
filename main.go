package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func SingleFizzBuzz(n int) string {
	if n%3 == 0 && n%5 == 0 {
		return "FizzBuzz"
	} else if n%3 == 0 {
		return "Fizz"
	} else if n%5 == 0 {
		return "Buzz"
	}
	return strconv.Itoa(n)
}

func handleRangeFizzBuzz(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := strconv.Atoi(fromStr)
	if err != nil {
		http.Error(w, "Invalid 'from' parameter", http.StatusBadRequest)
		return
	}
	to, err := strconv.Atoi(toStr)
	if err != nil {
		http.Error(w, "Invalid 'to' parameter", http.StatusBadRequest)
		return
	}
	if from > to {
		http.Error(w, "'from' should be less than or equal to 'to'", http.StatusBadRequest)
		return
	}
	if to-from > 100 {
		http.Error(w, "Range should not be greater than 100", http.StatusBadRequest)
		return
	}

	start := time.Now()
	var wg sync.WaitGroup
	results := make([]string, to-from+1)
	sem := make(chan struct{}, 1000)

	for i := from; i <= to; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			results[i-from] = SingleFizzBuzz(i)
			<-sem
		}(i)
	}

	wg.Wait()
	response := strings.Join(results, " ")

	latency := time.Since(start)
	log.Printf("Request: from=%d, to=%d, Response: %s, Latency: %s", from, to, response, latency)

	fmt.Fprintln(w, response)
}

func main() {
	srv := &http.Server{
		Addr:         ":9000",
		WriteTimeout: 1 * time.Second,
		ReadTimeout:  1 * time.Second,
		IdleTimeout:  30 * time.Second,
		Handler:      http.DefaultServeMux,
	}

	http.HandleFunc("/range-fizzbuzz", handleRangeFizzBuzz)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Server stopped: %s", err)
		}
	}()
	log.Println("Running server ... ")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Server gracefully stopped")
}
