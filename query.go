package hltb

import "strings"

type Query struct {
	term     string
	match    MatchType
	platform string
	modifier string
	page     int
	size     int
}

func NewQuery() *Query {
	return &Query{
		page: 1,
		size: 20,
	}
}

func (q *Query) SetTerm(term string, match MatchType) *Query {
	q.term = term
	q.match = match
	return q
}

func (q *Query) SetPlatform(platform string) *Query {
	q.platform = platform
	return q
}

func (q *Query) SetModifier(modifier string) *Query {
	q.modifier = modifier
	return q
}

func (q *Query) SetPage(page int) *Query {
	q.page = page
	return q
}

func (q *Query) SetSize(size int) *Query {
	q.size = size
	return q
}

func (q *Query) build() seekQuery {
	var terms []string
	if q.term != "" {
		if q.match == MatchFuzzy {
			terms = strings.Fields(q.term)
		} else {
			terms = []string{q.term}
		}
	} else {
		terms = []string{}
	}

	return seekQuery{
		SearchType:  "games",
		SearchTerms: terms,
		SearchPage:  q.page,
		Size:        q.size,
		SearchOptions: searchOptions{
			Games: gameOptions{
				Platform: q.platform,
				Modifier: q.modifier,
				RangeTime: rangeTime{
					Min: nil,
					Max: nil,
				},
				Gameplay: gameplay{
					Perspective: "",
					Flow:        "",
					Genre:       "",
					Difficulty:  "",
				},
			},
		},
		UseCache: true,
	}
}

func (q *Query) buildPayload(hpKey, hpVal string) any {
	built := q.build()
	if hpKey == "" || hpVal == "" {
		return built
	}

	return map[string]any{
		"searchType":    built.SearchType,
		"searchTerms":   built.SearchTerms,
		"searchPage":    built.SearchPage,
		"size":          built.Size,
		"searchOptions": built.SearchOptions,
		"useCache":      built.UseCache,
		hpKey:           hpVal,
	}
}

type seekQuery struct {
	SearchType    string        `json:"searchType"`
	SearchTerms   []string      `json:"searchTerms"`
	SearchPage    int           `json:"searchPage"`
	Size          int           `json:"size"`
	SearchOptions searchOptions `json:"searchOptions"`
	UseCache      bool          `json:"useCache"`
}

type searchOptions struct {
	Games gameOptions `json:"games"`
}

type gameOptions struct {
	Platform  string    `json:"platform"`
	Modifier  string    `json:"modifier"`
	RangeTime rangeTime `json:"rangeTime"`
	Gameplay  gameplay  `json:"gameplay"`
}

type rangeTime struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

type gameplay struct {
	Perspective string `json:"perspective"`
	Flow        string `json:"flow"`
	Genre       string `json:"genre"`
	Difficulty  string `json:"difficulty"`
}
