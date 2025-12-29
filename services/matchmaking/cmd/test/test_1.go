package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	fmt.Println("Hello, World!")
	tournamentID := "2c449397-96fe-442e-a2ee-e7684c4cb73b"
	url := fmt.Sprintf("%s/internal/tournaments/%s", "http://localhost:8089", tournamentID)

	// HTTP
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("tournament service request failed", err)
	}
	defer resp.Body.Close()

	fmt.Println(
		resp.Status,
		resp.Body,
	)

}
