// internal/entity/deck/deck.go
package deck

type Deck struct {
	ID         int64  `json:"id"`
	Name       string `json:"name" validate:"required,min=1,max=100"`
	Color      string `json:"color" validate:"required,oneof=W U B R G WU WB WR WG UB UR UG BR BG RG WUB WUR WUG WBR WBG WRG UBR UBG URG BRG WUBR WUBG WURG WBRG UBRG WUBRG"` // ex: "WUBRG"
	Commander  string `json:"commander" validate:"required,min=1,max=100"`
	OwnerID    int64  `json:"owner_id" validate:"required,gt=0"`
	SourceLink string `json:"source_link" validate:"omitempty,url"` // ex: https://archidekt.com/decks/123456
}
