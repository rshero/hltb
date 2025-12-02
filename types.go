package hltb

type GameType string

const (
	TypeGame        GameType = "game"
	TypeDLC         GameType = "dlc"
	TypeCompilation GameType = "compil"
)

type MatchType int

const (
	MatchExact MatchType = iota
	MatchFuzzy
)

type Game struct {
	ID            uint64   `json:"id"`
	Title         string   `json:"title"`
	Type          GameType `json:"type"`
	ImageURL      string   `json:"image_url"`
	MainStory     float32  `json:"main_story"`
	MainPlusExtra float32  `json:"main_plus_extra"`
	Completionist float32  `json:"completionist"`
	Platforms     []string `json:"platforms,omitempty"`
	SteamAppID    uint64   `json:"steam_app_id,omitempty"`
}

func parseGameType(s string) GameType {
	switch s {
	case "dlc":
		return TypeDLC
	case "compil":
		return TypeCompilation
	default:
		return TypeGame
	}
}
