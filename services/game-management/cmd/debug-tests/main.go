package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Game struct {
	ID          string  `json:"id"`
	Slug        string  `json:"slug"`
	Name        string  `json:"name"`
	StoragePath string  `json:"storagePath"`
	Version     string  `json:"version"`
	IsActive    bool    `json:"isActive"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	Description *string `json:"description,omitempty"`
	LogoURL     *string `json:"logoUrl,omitempty"`
}

type GamesListResponse struct {
	Games []Game `json:"games"`
	Total int    `json:"total"`
}

func main() {
	// URL to test - using direct service port
	url := "http://localhost:8085/v1/games"

	fmt.Println("🔍 Testing GET", url)
	fmt.Println("=====================================")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Error reading response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("📊 Status Code: %d\n", resp.StatusCode)
	fmt.Println("=====================================")

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("❌ Non-OK status code: %d\n", resp.StatusCode)
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}

	// Parse JSON response
	var gamesResp GamesListResponse
	if err := json.Unmarshal(body, &gamesResp); err != nil {
		fmt.Printf("❌ Error parsing JSON: %v\n", err)
		fmt.Printf("Response body: %s\n", string(body))
		os.Exit(1)
	}

	// Print results
	fmt.Printf("✅ Successfully fetched %d game(s)\n\n", gamesResp.Total)

	for i, game := range gamesResp.Games {
		fmt.Printf("Game #%d:\n", i+1)
		fmt.Printf("  ID:          %s\n", game.ID)
		fmt.Printf("  Slug:        %s\n", game.Slug)
		fmt.Printf("  Name:        %s\n", game.Name)
		fmt.Printf("  Version:     %s\n", game.Version)
		fmt.Printf("  StoragePath: %s", game.StoragePath)
		if game.StoragePath == "" {
			fmt.Printf(" ❌ EMPTY!\n")
		} else {
			fmt.Printf(" ✅\n")
		}
		fmt.Printf("  IsActive:    %v\n", game.IsActive)
		fmt.Printf("  CreatedAt:   %s\n", game.CreatedAt)
		fmt.Printf("  UpdatedAt:   %s\n", game.UpdatedAt)
		if game.Description != nil {
			fmt.Printf("  Description: %s\n", *game.Description)
		}
		if game.LogoURL != nil {
			fmt.Printf("  LogoURL:     %s\n", *game.LogoURL)
		}
		fmt.Println()
	}

	// Print raw JSON for verification
	fmt.Println("=====================================")
	fmt.Println("📝 Raw JSON Response:")
	fmt.Println("=====================================")
	prettyJSON, _ := json.MarshalIndent(gamesResp, "", "  ")
	fmt.Println(string(prettyJSON))
}
