package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	deckEntity "github.com/josofm/liliana/internal/entity/deck"
)

const scryfallBatchSize = 75

type CardValidator interface {
	Validate(cards []deckEntity.Card) ([]deckEntity.Card, error)
	ResolveCommander(name string) (deckEntity.Card, error)
}

type ScryfallValidator struct {
	client      *http.Client
	baseURL     string
	mu          sync.Mutex
	lastRequest time.Time
	cache       map[string]deckEntity.Card
}

func NewScryfallValidator() *ScryfallValidator {
	return NewScryfallValidatorWithBaseURL(&http.Client{Timeout: 10 * time.Second}, "https://api.scryfall.com")
}

func NewScryfallValidatorWithBaseURL(client *http.Client, baseURL string) *ScryfallValidator {
	return &ScryfallValidator{client: client, baseURL: strings.TrimRight(baseURL, "/"), cache: make(map[string]deckEntity.Card)}
}

type scryfallIdentifier struct {
	Name string `json:"name"`
}
type scryfallCollectionRequest struct {
	Identifiers []scryfallIdentifier `json:"identifiers"`
}
type scryfallCollectionResponse struct {
	Data     []scryfallCard       `json:"data"`
	NotFound []scryfallIdentifier `json:"not_found"`
}
type scryfallCard struct {
	OracleID      string            `json:"oracle_id"`
	Name          string            `json:"name"`
	ManaCost      string            `json:"mana_cost"`
	TypeLine      string            `json:"type_line"`
	OracleText    string            `json:"oracle_text"`
	ColorIdentity []string          `json:"color_identity"`
	ImageURIs     map[string]string `json:"image_uris"`
	CardFaces     []struct {
		ImageURIs map[string]string `json:"image_uris"`
	} `json:"card_faces"`
}

func (v *ScryfallValidator) ResolveCommander(name string) (deckEntity.Card, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if wait := 100*time.Millisecond - time.Since(v.lastRequest); wait > 0 {
		time.Sleep(wait)
	}
	req, err := http.NewRequest(http.MethodGet, v.baseURL+"/cards/named?exact="+url.QueryEscape(name), nil)
	if err != nil {
		return deckEntity.Card{}, err
	}
	req.Header.Set("Accept", "application/json;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "liliana/1.0 (https://github.com/josofm/liliana)")
	resp, err := v.client.Do(req)
	v.lastRequest = time.Now()
	if err != nil {
		return deckEntity.Card{}, fmt.Errorf("find commander with Scryfall: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return deckEntity.Card{}, fmt.Errorf("commander not found: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return deckEntity.Card{}, fmt.Errorf("find commander with Scryfall: status %d", resp.StatusCode)
	}
	var source scryfallCard
	if err := json.NewDecoder(resp.Body).Decode(&source); err != nil {
		return deckEntity.Card{}, fmt.Errorf("decode Scryfall commander response: %w", err)
	}
	if !strings.EqualFold(source.Name, name) {
		return deckEntity.Card{}, fmt.Errorf("commander not found: %s", name)
	}
	if !canBeCommander(source) {
		return deckEntity.Card{}, fmt.Errorf("card cannot be a commander: %s", source.Name)
	}
	return cardFromScryfall(source), nil
}

func canBeCommander(card scryfallCard) bool {
	return strings.Contains(card.TypeLine, "Legendary Creature") ||
		strings.Contains(strings.ToLower(card.OracleText), "can be your commander")
}

func cardFromScryfall(source scryfallCard) deckEntity.Card {
	imageURI := source.ImageURIs["normal"]
	if imageURI == "" && len(source.CardFaces) > 0 {
		imageURI = source.CardFaces[0].ImageURIs["normal"]
	}
	return deckEntity.Card{OracleID: source.OracleID, Name: source.Name, ManaCost: source.ManaCost, TypeLine: source.TypeLine, ColorIdentity: source.ColorIdentity, ImageURI: imageURI}
}

func (v *ScryfallValidator) Validate(cards []deckEntity.Card) ([]deckEntity.Card, error) {
	result := make([]deckEntity.Card, len(cards))
	copy(result, cards)
	pending := make([]int, 0, len(cards))
	for index, card := range result {
		if card.OracleID != "" {
			continue
		}
		key := strings.ToLower(card.Name)
		v.mu.Lock()
		cached, ok := v.cache[key]
		v.mu.Unlock()
		if ok {
			cached.Quantity = card.Quantity
			result[index] = cached
			continue
		}
		pending = append(pending, index)
	}

	for start := 0; start < len(pending); start += scryfallBatchSize {
		end := min(start+scryfallBatchSize, len(pending))
		indices := pending[start:end]
		found, notFound, err := v.fetch(result, indices)
		if err != nil {
			return nil, err
		}
		if len(notFound) > 0 {
			return nil, fmt.Errorf("cards not found: %s", strings.Join(notFound, ", "))
		}
		for _, index := range indices {
			card, ok := found[strings.ToLower(result[index].Name)]
			if !ok {
				return nil, fmt.Errorf("card not found: %s", result[index].Name)
			}
			card.Quantity = result[index].Quantity
			result[index] = card
		}
	}
	return result, nil
}

func (v *ScryfallValidator) fetch(cards []deckEntity.Card, indices []int) (map[string]deckEntity.Card, []string, error) {
	payload := scryfallCollectionRequest{Identifiers: make([]scryfallIdentifier, len(indices))}
	for position, index := range indices {
		payload.Identifiers[position] = scryfallIdentifier{Name: cards[index].Name}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, err
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	if wait := 100*time.Millisecond - time.Since(v.lastRequest); wait > 0 {
		time.Sleep(wait)
	}
	req, err := http.NewRequest(http.MethodPost, v.baseURL+"/cards/collection", bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "liliana/1.0 (https://github.com/josofm/liliana)")
	resp, err := v.client.Do(req)
	v.lastRequest = time.Now()
	if err != nil {
		return nil, nil, fmt.Errorf("validate cards with Scryfall: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("validate cards with Scryfall: status %d", resp.StatusCode)
	}
	var response scryfallCollectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, nil, fmt.Errorf("decode Scryfall response: %w", err)
	}

	found := make(map[string]deckEntity.Card, len(response.Data))
	for _, source := range response.Data {
		card := cardFromScryfall(source)
		found[strings.ToLower(source.Name)] = card
		v.cache[strings.ToLower(source.Name)] = card
	}
	notFound := make([]string, len(response.NotFound))
	for index, identifier := range response.NotFound {
		notFound[index] = identifier.Name
	}
	return found, notFound, nil
}
