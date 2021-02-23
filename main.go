package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

// QuoteData describes a quote
type QuoteData struct {
	ID         string   `json:"id"`
	Quote      string   `json:"quote"`
	Length     string   `json:"length"`
	Author     string   `json:"author"`
	Tags       []string `json:"tags"`
	Category   string   `json:"category"`
	Date       string   `json:"date"`
	Permalink  string   `json:"permalink"`
	Title      string   `json:"title"`
	Background string   `json:"background"`
}

// APISuccess describes a successful API response
type APISuccess struct {
	Total string `json:"total"`
}

// QuoteContent describes the payload of a successful API response
type QuoteContent struct {
	Quotes    []QuoteData `json:"quotes"`
	Copyright string      `json:"copyright"`
}

// QuoteResponse describes an API response
type QuoteResponse struct {
	Success  APISuccess   `json:"success"`
	Contents QuoteContent `json:"contents"`
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome! Please hit the `/qod` API to get the quote of the day."))
}

func getQuoteFromAPI() (*QuoteResponse, error) {
	resp, err := http.Get("http://quotes.rest/qod.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	log.Println("Quote API returned:", resp.StatusCode, http.StatusText(resp.StatusCode))

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		quoteResp := &QuoteResponse{}
		json.NewDecoder(resp.Body).Decode(quoteResp)
		return quoteResp, nil
	}
	return nil, errors.New("Could not get quote from API")
}

func quoteOfTheDayHandler(client *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentTime := time.Now()
		date := currentTime.Format("2006-01-02")

		val, err := client.Get(date).Result()
		if err == redis.Nil {
			log.Println("Cache miss for date", date)
			quoteResp, err := getQuoteFromAPI()
			if err != nil {
				w.Write([]byte("Sorry! We could not get the Quote of the Day. Please try again."))
				return
			}
			quote := quoteResp.Contents.Quotes[0].Quote
			client.Set(date, quote, 24*time.Hour)
			w.Write([]byte(quote))
		} else {
			log.Println("Cache hit for date", date)
			w.Write([]byte(val))
		}
	}
}

func waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// block until an interrupt is received
	<-interruptChan

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

func main() {
	var (
		host     = getEnv("REDIS_HOST", "localhost")
		port     = string(getEnv("REDIS_PORT", "6379"))
		password = getEnv("REDIS_PASSWORD", "")
	)

	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/qod", quoteOfTheDayHandler(client))

	srv := &http.Server{
		Handler:      r,
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Starting server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	waitForShutdown(srv)
}
