package hltb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	baseURL   = "https://howlongtobeat.com"
	userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0"
)

var (
	ErrNoResults = errors.New("no results found")
	steamIDRegex = regexp.MustCompile(`store\.steampowered\.com/app/(\d+)`)
)

type Client struct {
	http      *http.Client
	authToken string
	mu        sync.RWMutex
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 5,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func NewClientWithInit() (*Client, error) {
	c := NewClient()
	if err := c.refreshToken(); err != nil {
		return nil, fmt.Errorf("failed to initialize: %w", err)
	}
	return c, nil
}

func (c *Client) Search(q *Query) ([]Game, error) {
	if err := c.ensureToken(); err != nil {
		return nil, err
	}
	return c.search(q, true)
}

func (c *Client) SearchByName(name string) ([]Game, error) {
	return c.Search(NewQuery().SetTerm(name, MatchFuzzy))
}

func (c *Client) SearchFirst(name string) (*Game, error) {
	games, err := c.SearchByName(name)
	if err != nil {
		return nil, err
	}
	if len(games) == 0 {
		return nil, ErrNoResults
	}
	return &games[0], nil
}

func (c *Client) SearchFirstWithDetails(name string) (*Game, error) {
	game, err := c.SearchFirst(name)
	if err != nil {
		return nil, err
	}
	return c.FetchDetails(game)
}

func (c *Client) FetchDetails(game *Game) (*Game, error) {
	if game == nil {
		return nil, errors.New("game is nil")
	}

	url := fmt.Sprintf("%s/game/%d", baseURL, game.ID)
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch details failed [%d]", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	html := string(body)

	if match := steamIDRegex.FindStringSubmatch(html); len(match) > 1 {
		var steamID uint64
		fmt.Sscanf(match[1], "%d", &steamID)
		game.SteamAppID = steamID
	}

	game.Platforms = extractPlatforms(html)

	return game, nil
}

func extractPlatforms(html string) []string {
	var platforms []string

	// Find "Platform" or "Platforms" followed by ":</strong>"
	idx := strings.Index(html, "Platform")
	if idx == -1 {
		return platforms
	}

	// Find the closing </strong> tag after Platform
	strongEnd := strings.Index(html[idx:], ":</strong>")
	if strongEnd == -1 {
		return platforms
	}

	start := idx + strongEnd + len(":</strong>")
	rest := html[start:]

	// Skip <br/> or <br> if present
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "<br") {
		if brEnd := strings.Index(rest, ">"); brEnd != -1 {
			rest = rest[brEnd+1:]
		}
	}

	// Find end - either </div> or <
	end := strings.Index(rest, "</")
	if end == -1 {
		end = strings.Index(rest, "<")
	}
	if end == -1 {
		return platforms
	}

	platformStr := strings.TrimSpace(rest[:end])
	if platformStr == "" {
		return platforms
	}

	parts := strings.Split(platformStr, ",")
	for _, p := range parts {
		platform := strings.TrimSpace(p)
		if platform != "" {
			platforms = append(platforms, platform)
		}
	}

	return platforms
}

func (c *Client) search(q *Query, retry bool) ([]Game, error) {
	body, err := json.Marshal(q.build())
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	req, err := c.newRequest("POST", baseURL+"/api/finder", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden && retry {
		if err := c.refreshToken(); err != nil {
			return nil, err
		}
		return c.search(q, false)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed [%d]: %s", resp.StatusCode, body)
	}

	var result apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	games := make([]Game, len(result.Data))
	for i, item := range result.Data {
		games[i] = Game{
			ID:            item.GameID,
			Title:         item.GameName,
			Type:          parseGameType(item.GameType),
			ImageURL:      fmt.Sprintf("%s/games/%s", baseURL, item.GameImage),
			MainStory:     secsToHours(item.CompMain),
			MainPlusExtra: secsToHours(item.CompPlus),
			Completionist: secsToHours(item.Comp100),
		}
	}

	return games, nil
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Origin", baseURL)
	req.Header.Set("Referer", baseURL)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.mu.RLock()
	if c.authToken != "" {
		req.Header.Set("x-auth-token", c.authToken)
	}
	c.mu.RUnlock()

	return req, nil
}

func (c *Client) refreshToken() error {
	url := fmt.Sprintf("%s/api/finder/init?t=%d", baseURL, time.Now().UnixMilli())

	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed [%d]", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	c.mu.Lock()
	c.authToken = result.Token
	c.mu.Unlock()

	return nil
}

func (c *Client) ensureToken() error {
	c.mu.RLock()
	hasToken := c.authToken != ""
	c.mu.RUnlock()

	if !hasToken {
		return c.refreshToken()
	}
	return nil
}

type apiResponse struct {
	Data []struct {
		GameID    uint64 `json:"game_id"`
		GameName  string `json:"game_name"`
		GameType  string `json:"game_type"`
		GameImage string `json:"game_image"`
		CompMain  uint32 `json:"comp_main"`
		CompPlus  uint32 `json:"comp_plus"`
		Comp100   uint32 `json:"comp_100"`
	} `json:"data"`
}

func secsToHours(secs uint32) float32 {
	if secs == 0 {
		return 0
	}
	return float32(secs) / 3600
}
