// internal/entity/deck/deck.go
package deck

type Deck struct {
	ID                int64  `json:"id"`
	Name              string `json:"name" validate:"required,min=1,max=100"`
	Color             string `json:"color" validate:"required,oneof=C W U B R G WU WB WR WG UB UR UG BR BG RG WUB WUR WUG WBR WBG WRG UBR UBG URG BRG WUBR WUBG WURG WBRG UBRG WUBRG"` // ex: "WUBRG"
	Format            string `json:"format" validate:"required,oneof=commander standard modern pioneer legacy vintage pauper brawl oathbreaker limited"`
	Commander         string `json:"commander" validate:"required_if=Format commander,omitempty,min=1,max=100"`
	CommanderImageURI string `json:"commander_image_uri" validate:"omitempty,url"`
	OwnerID           int64  `json:"owner_id" validate:"required,gt=0"`
	SourceLink        string `json:"source_link" validate:"omitempty,url"` // ex: https://archidekt.com/decks/123456
	Cards             []Card `json:"cards"`
}

type Card struct {
	OracleID      string   `json:"oracle_id"`
	Name          string   `json:"name"`
	Quantity      int      `json:"quantity"`
	ManaCost      string   `json:"mana_cost,omitempty"`
	TypeLine      string   `json:"type_line,omitempty"`
	ColorIdentity []string `json:"color_identity,omitempty"`
	ImageURI      string   `json:"image_uri,omitempty"`
}
