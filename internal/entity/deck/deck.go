// internal/entity/deck/deck.go
package deck

type Deck struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"` // ex: "WUBRG"
	Commander  string `json:"commander"`
	OwnerID    int64  `json:"owner_id"`
	SourceLink string `json:"source_link"` // ex: https://archidekt.com/decks/123456
}
