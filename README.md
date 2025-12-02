# HLTB - HowLongToBeat Go Client

A fast, lightweight Go client for fetching game completion times from [HowLongToBeat.com](https://howlongtobeat.com).

## Installation

```bash
go get github.com/rshero/hltb
```

## Usage

### Quick Start

```go
package main

import (
    "fmt"
    "github.com/rshero/hltb"
)

func main() {
    client, err := hltb.NewClientWithInit()
    if err != nil {
        panic(err)
    }

    game, err := client.SearchFirst("Elden Ring")
    if err != nil {
        panic(err)
    }

    fmt.Printf("%s - Main: %.1fh, Extra: %.1fh, 100%%: %.1fh\n",
        game.Title, game.MainStory, game.MainPlusExtra, game.Completionist)
}
```

### Search with Full Details

Fetch platforms and Steam App ID by fetching the game's detail page:

```go
game, err := client.SearchFirstWithDetails("Elden Ring")
if err != nil {
    panic(err)
}

fmt.Printf("Title: %s\n", game.Title)
fmt.Printf("Steam ID: %d\n", game.SteamAppID)
fmt.Printf("Platforms: %v\n", game.Platforms)
fmt.Printf("Main Story: %.1f hours\n", game.MainStory)
```

Output:
```
Title: Elden Ring
Steam ID: 1245620
Platforms: [Nintendo Switch 2 PC PlayStation 4 PlayStation 5 Xbox One Xbox Series X/S]
Main Story: 59.9 hours
```

### Search with Filters

```go
query := hltb.NewQuery().
    SetTerm("Zelda", hltb.MatchFuzzy).
    SetPlatform(hltb.PlatformNintendoSwitch).
    SetModifier(hltb.ModifierHideDLC)

games, err := client.Search(query)
```

### Fetch Details for Existing Game

If you already have a game from search results, you can fetch additional details:

```go
game, _ := client.SearchFirst("Metal Gear")
game, _ = client.FetchDetails(game) // Adds Platforms and SteamAppID
```

## API Reference

### Client Methods

| Method | Description |
|--------|-------------|
| `NewClient()` | Creates a new client |
| `NewClientWithInit()` | Creates client with pre-fetched auth token (recommended) |
| `Search(query)` | Search with a custom query |
| `SearchByName(name)` | Search by game name |
| `SearchFirst(name)` | Get first result for a game name |
| `SearchFirstWithDetails(name)` | Get first result with platforms and Steam ID |
| `FetchDetails(game)` | Fetch platforms and Steam ID for a game |

### Query Builder

| Method | Description |
|--------|-------------|
| `SetTerm(term, matchType)` | Set search term (`MatchExact` or `MatchFuzzy`) |
| `SetPlatform(platform)` | Filter by platform |
| `SetModifier(modifier)` | Set result modifier |
| `SetPage(page)` | Set page number |
| `SetSize(size)` | Set results per page |

### Game Struct

```go
type Game struct {
    ID            uint64   // HowLongToBeat ID
    Title         string   // Game title
    Type          GameType // game, dlc, or compil
    ImageURL      string   // Cover image URL
    MainStory     float32  // Hours to complete main story
    MainPlusExtra float32  // Hours for main + extras
    Completionist float32  // Hours for 100% completion
    Platforms     []string // Available platforms (requires FetchDetails)
    SteamAppID    uint64   // Steam App ID (requires FetchDetails)
}
```

### Platforms

```go
hltb.PlatformPC
hltb.PlatformPlayStation5
hltb.PlatformPlayStation4
hltb.PlatformXboxSeriesXS
hltb.PlatformXboxOne
hltb.PlatformNintendoSwitch
hltb.PlatformNintendo3DS
// ... see platforms.go for full list
```

### Modifiers

```go
hltb.ModifierNone      // No filter
hltb.ModifierHideDLC   // Hide DLC entries
hltb.ModifierOnlyDLC   // Show only DLC
hltb.ModifierOnlyMods  // Show only mods
hltb.ModifierOnlyHacks // Show only hacks
```

## Performance

| Operation | Typical Time |
|-----------|--------------|
| `SearchFirst` | 700-900ms |
| `SearchFirstWithDetails` | 1.5-2s |
| `FetchDetails` | 500-800ms |

The client is optimized with:
- Pre-fetched auth tokens
- HTTP connection reuse (keep-alive)
- Thread-safe for concurrent use

## Running Tests

```bash
# Unit tests only
go test -short

# All tests including integration
go test -v

# Run specific test
go test -v -run TestSearchFirstWithDetails

# Benchmarks
go test -bench=. -benchtime=5x -run=^$
```

## License

MIT