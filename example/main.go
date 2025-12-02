package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rshero/hltb"
)

func main() {
	client, err := hltb.NewClientWithInit()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Search with full details (includes platforms and Steam ID)
	fmt.Println("=== Search with Details ===")
	start := time.Now()
	game, err := client.SearchFirstWithDetails("Absolum")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	data, _ := json.MarshalIndent(game, "", "  ")
	fmt.Println(string(data))
	fmt.Printf("Completed in %v\n\n", time.Since(start))

	// Another example with a popular game
	fmt.Println("=== Elden Ring Details ===")
	start = time.Now()
	game, err = client.SearchFirstWithDetails("Elden Ring")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	data, _ = json.MarshalIndent(game, "", "  ")
	fmt.Println(string(data))
	fmt.Printf("Completed in %v\n\n", time.Since(start))

	// Simple search without details (faster)
	fmt.Println("=== Simple Search (no details) ===")
	start = time.Now()
	game, err = client.SearchFirst("Metal Gear")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	data, _ = json.MarshalIndent(game, "", "  ")
	fmt.Println(string(data))
	fmt.Printf("Completed in %v\n\n", time.Since(start))

	// Search with platform filter
	fmt.Println("=== Zelda on Switch ===")
	start = time.Now()
	query := hltb.NewQuery().
		SetTerm("Zelda", hltb.MatchFuzzy).
		SetPlatform(hltb.PlatformNintendoSwitch)

	games, err := client.Search(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for i, g := range games {
		if i >= 5 {
			break
		}
		fmt.Printf("%d. %s - Main: %.1fh, Extra: %.1fh, 100%%: %.1fh\n",
			i+1, g.Title, g.MainStory, g.MainPlusExtra, g.Completionist)
	}
	fmt.Printf("Completed in %v\n", time.Since(start))
}
