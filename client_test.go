package hltb

import (
	"sync"
	"testing"
	"time"
)

var (
	testClient     *Client
	testClientOnce sync.Once
	testClientErr  error
)

func getTestClient(t *testing.T) *Client {
	testClientOnce.Do(func() {
		testClient, testClientErr = NewClientWithInit()
	})
	if testClientErr != nil {
		t.Fatalf("failed to create test client: %v", testClientErr)
	}
	return testClient
}

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
	if c.http == nil {
		t.Error("http client is nil")
	}
}

func TestNewClientWithInit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c, err := NewClientWithInit()
	if err != nil {
		t.Fatalf("NewClientWithInit failed: %v", err)
	}
	if c.authToken == "" {
		t.Error("auth token should be set after init")
	}
}

func TestSearchByName(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	games, err := c.SearchByName("Metal Gear")
	if err != nil {
		t.Fatalf("SearchByName failed: %v", err)
	}
	if len(games) == 0 {
		t.Fatal("expected results")
	}

	found := false
	for _, g := range games {
		if g.Title == "Metal Gear" {
			found = true
			if g.ID == 0 {
				t.Error("game ID should not be 0")
			}
			if g.MainStory <= 0 {
				t.Error("main story time should be > 0")
			}
			break
		}
	}
	if !found {
		t.Log("'Metal Gear' not in top results, but search succeeded")
	}
}

func TestSearchFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	game, err := c.SearchFirst("Metal Gear")
	if err != nil {
		t.Fatalf("SearchFirst failed: %v", err)
	}
	if game == nil {
		t.Fatal("expected a game")
	}
	if game.ID == 0 {
		t.Error("game ID should not be 0")
	}
	if game.Title == "" {
		t.Error("game title should not be empty")
	}

	t.Logf("Found: %s (ID: %d) - Main: %.2fh, Extra: %.2fh, 100%%: %.2fh",
		game.Title, game.ID, game.MainStory, game.MainPlusExtra, game.Completionist)
}

func TestSearchFirstWithDetails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	game, err := c.SearchFirstWithDetails("Elden Ring")
	if err != nil {
		t.Fatalf("SearchFirstWithDetails failed: %v", err)
	}
	if game == nil {
		t.Fatal("expected a game")
	}

	if game.SteamAppID == 0 {
		t.Error("expected steam app id")
	}
	if len(game.Platforms) == 0 {
		t.Error("expected platforms")
	}

	t.Logf("Game: %s", game.Title)
	t.Logf("Steam ID: %d", game.SteamAppID)
	t.Logf("Platforms: %v", game.Platforms)
}

func TestFetchDetails(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)

	// First get a game without details
	game, err := c.SearchFirst("Absolum")
	if err != nil {
		t.Fatalf("SearchFirst failed: %v", err)
	}

	if game.SteamAppID != 0 {
		t.Error("steam app id should be 0 before fetch details")
	}
	if len(game.Platforms) != 0 {
		t.Error("platforms should be empty before fetch details")
	}

	// Now fetch details
	game, err = c.FetchDetails(game)
	if err != nil {
		t.Fatalf("FetchDetails failed: %v", err)
	}

	if game.SteamAppID == 0 {
		t.Error("expected steam app id after fetch details")
	}
	if len(game.Platforms) == 0 {
		t.Error("expected platforms after fetch details")
	}

	t.Logf("Game: %s", game.Title)
	t.Logf("Steam ID: %d", game.SteamAppID)
	t.Logf("Platforms: %v", game.Platforms)
}

func TestSearchWithPlatform(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	q := NewQuery().
		SetTerm("Zelda", MatchFuzzy).
		SetPlatform(PlatformNintendoSwitch)

	games, err := c.Search(q)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(games) == 0 {
		t.Fatal("expected results")
	}

	t.Logf("Found %d Zelda games for Switch", len(games))
	for i, g := range games {
		if i >= 3 {
			break
		}
		t.Logf("  %d. %s", i+1, g.Title)
	}
}

func TestSearchIterations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	iterations := 5
	var total time.Duration

	t.Logf("Running %d search iterations...", iterations)

	for i := range iterations {
		start := time.Now()
		game, err := c.SearchFirst("Metal Gear")
		elapsed := time.Since(start)
		total += elapsed

		if err != nil {
			t.Errorf("iteration %d failed: %v", i+1, err)
			continue
		}
		t.Logf("  %d: %v - %s", i+1, elapsed, game.Title)
	}

	avg := total / time.Duration(iterations)
	t.Logf("Total: %v | Avg: %v", total, avg)
}

func TestErrNoResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	c := getTestClient(t)
	_, err := c.SearchFirst("xyznonexistentgame12345xyz")
	if err != ErrNoResults {
		t.Errorf("expected ErrNoResults, got: %v", err)
	}
}

func BenchmarkSearchFirst(b *testing.B) {
	c, err := NewClientWithInit()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := c.SearchFirst("Metal Gear")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNewQuery(b *testing.B) {
	for b.Loop() {
		NewQuery().
			SetTerm("Metal Gear Solid", MatchFuzzy).
			SetPlatform(PlatformPC).
			SetModifier(ModifierHideDLC)
	}
}
