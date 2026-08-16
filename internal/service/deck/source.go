package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
)

var archidektDeckPath = regexp.MustCompile(`^/decks/(\d+)(?:/.*)?$`)

var ErrUnsupportedSource = errors.New("unsupported deck source")

type SourceImporter interface {
	Import(sourceLink string) (*deckEntity.Deck, error)
}

type ArchidektImporter struct {
	client  *http.Client
	baseURL string
}

func NewArchidektImporter() *ArchidektImporter {
	return &ArchidektImporter{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://archidekt.com",
	}
}

func NewArchidektImporterWithBaseURL(client *http.Client, baseURL string) *ArchidektImporter {
	return &ArchidektImporter{client: client, baseURL: strings.TrimRight(baseURL, "/")}
}

type archidektResponse struct {
	Name       string              `json:"name"`
	DeckFormat int                 `json:"deckFormat"`
	Categories []archidektCategory `json:"categories"`
	Cards      []archidektDeckCard `json:"cards"`
}

type archidektCategory struct {
	Name           string `json:"name"`
	IncludedInDeck bool   `json:"includedInDeck"`
}

type archidektDeckCard struct {
	Categories []string `json:"categories"`
	Quantity   int      `json:"quantity"`
	Card       struct {
		OracleCard struct {
			Name          string   `json:"name"`
			UID           string   `json:"uid"`
			ManaCost      string   `json:"manaCost"`
			ColorIdentity []string `json:"colorIdentity"`
			SuperTypes    []string `json:"superTypes"`
			Types         []string `json:"types"`
			SubTypes      []string `json:"subTypes"`
		} `json:"oracleCard"`
	} `json:"card"`
}

func (i *ArchidektImporter) Import(sourceLink string) (*deckEntity.Deck, error) {
	u, err := url.Parse(sourceLink)
	if err != nil || !strings.EqualFold(u.Hostname(), "archidekt.com") {
		return nil, ErrUnsupportedSource
	}

	match := archidektDeckPath.FindStringSubmatch(u.Path)
	if len(match) != 2 {
		return nil, ErrUnsupportedSource
	}

	req, err := http.NewRequest(http.MethodGet, i.baseURL+"/api/decks/"+match[1]+"/", nil)
	if err != nil {
		return nil, fmt.Errorf("build Archidekt request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "liliana/1.0")

	resp, err := i.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch Archidekt deck: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch Archidekt deck: status %d", resp.StatusCode)
	}

	var source archidektResponse
	if err := json.NewDecoder(resp.Body).Decode(&source); err != nil {
		return nil, fmt.Errorf("decode Archidekt deck: %w", err)
	}

	includedCategories := make(map[string]bool)
	for _, category := range source.Categories {
		includedCategories[category.Name] = category.IncludedInDeck
	}

	colors := make(map[string]bool)
	cards := make([]deckEntity.Card, 0, len(source.Cards))
	commanders := make([]string, 0, 2)
	for _, sourceCard := range source.Cards {
		if !cardIsIncluded(sourceCard.Categories, includedCategories) || sourceCard.Quantity <= 0 || sourceCard.Card.OracleCard.Name == "" {
			continue
		}
		card := deckEntity.Card{
			Name: sourceCard.Card.OracleCard.Name, Quantity: sourceCard.Quantity,
			OracleID: sourceCard.Card.OracleCard.UID, ManaCost: sourceCard.Card.OracleCard.ManaCost,
			TypeLine:      archidektTypeLine(sourceCard.Card.OracleCard.SuperTypes, sourceCard.Card.OracleCard.Types, sourceCard.Card.OracleCard.SubTypes),
			ColorIdentity: sourceCard.Card.OracleCard.ColorIdentity,
		}
		cards = append(cards, card)
		for _, color := range sourceCard.Card.OracleCard.ColorIdentity {
			colors[color] = true
		}
		if containsCategory(sourceCard.Categories, "Commander") {
			commanders = append(commanders, card.Name)
		}
	}

	cards = mergeCardsByOracleID(cards)
	sort.Slice(cards, func(a, b int) bool { return cards[a].Name < cards[b].Name })
	return &deckEntity.Deck{
		Name:      source.Name,
		Color:     colorCode(colors),
		Format:    archidektFormat(source.DeckFormat),
		Commander: strings.Join(commanders, " / "),
		Cards:     cards,
	}, nil
}

func mergeCardsByOracleID(cards []deckEntity.Card) []deckEntity.Card {
	result := make([]deckEntity.Card, 0, len(cards))
	positions := make(map[string]int, len(cards))
	for _, card := range cards {
		key := card.OracleID
		if key == "" {
			key = strings.ToLower(card.Name)
		}
		if position, exists := positions[key]; exists {
			result[position].Quantity += card.Quantity
			continue
		}
		positions[key] = len(result)
		result = append(result, card)
	}
	return result
}

func archidektTypeLine(superTypes, types, subTypes []string) string {
	left := strings.Join(append(append([]string{}, superTypes...), types...), " ")
	if len(subTypes) == 0 {
		return left
	}
	return left + " — " + strings.Join(subTypes, " ")
}

func cardIsIncluded(categories []string, included map[string]bool) bool {
	for _, category := range categories {
		if included[category] {
			return true
		}
	}
	return false
}

func containsCategory(categories []string, want string) bool {
	for _, category := range categories {
		if strings.EqualFold(category, want) {
			return true
		}
	}
	return false
}

func colorCode(colors map[string]bool) string {
	var result strings.Builder
	for _, color := range []struct{ name, code string }{{"White", "W"}, {"Blue", "U"}, {"Black", "B"}, {"Red", "R"}, {"Green", "G"}} {
		if colors[color.name] {
			result.WriteString(color.code)
		}
	}
	if result.Len() == 0 {
		return "C"
	}
	return result.String()
}

func archidektFormat(value int) string {
	formats := map[int]string{1: "standard", 2: "modern", 3: "commander", 4: "legacy", 5: "vintage", 6: "pauper", 9: "brawl", 13: "pioneer", 15: "oathbreaker"}
	return formats[value]
}
