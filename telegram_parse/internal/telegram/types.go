package telegram

// Export represents the main structure of Telegram export JSON
type Export struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	ID       int64     `json:"id"`
	Messages []Message `json:"messages"`
}

// Message represents a single message in Telegram export
type Message struct {
	ID                  int64        `json:"id"`
	Type                string       `json:"type"`
	Date                string       `json:"date"`
	DateUnixtime        string       `json:"date_unixtime"`
	From                string       `json:"from,omitempty"`
	FromID              string       `json:"from_id,omitempty"`
	Text                interface{}  `json:"text"` // Can be string or []TextEntity
	TextEntities        []TextEntity `json:"text_entities,omitempty"`
	Photo               string       `json:"photo,omitempty"`
	File                string       `json:"file,omitempty"`
	Thumbnail           string       `json:"thumbnail,omitempty"`
	MediaType           string       `json:"media_type,omitempty"`
	MimeType            string       `json:"mime_type,omitempty"`
	Duration            int          `json:"duration,omitempty"`
	Width               int          `json:"width,omitempty"`
	Height              int          `json:"height,omitempty"`
	ForwardedFrom       string       `json:"forwarded_from,omitempty"`
	ReplyToMessageID    int64        `json:"reply_to_message_id,omitempty"`
	ViaBot              string       `json:"via_bot,omitempty"`
	Actor               string       `json:"actor,omitempty"`
	Action              string       `json:"action,omitempty"`
	Title               string       `json:"title,omitempty"`
	Members             []string     `json:"members,omitempty"`
	Inviter             string       `json:"inviter,omitempty"`
	Poll                *Poll        `json:"poll,omitempty"`
	ContactInformation  *Contact     `json:"contact_information,omitempty"`
	LocationInformation *Location    `json:"location_information,omitempty"`
	LiveLocationPeriod  int          `json:"live_location_period,omitempty"`
}

// TextEntity represents text formatting entity
type TextEntity struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Href   string `json:"href,omitempty"`
	UserID string `json:"user_id,omitempty"`
}

// Poll represents a poll message
type Poll struct {
	Question    string       `json:"question"`
	Closed      bool         `json:"closed"`
	TotalVoters int          `json:"total_voters"`
	Answers     []PollAnswer `json:"answers"`
}

// PollAnswer represents a poll answer option
type PollAnswer struct {
	Text   string `json:"text"`
	Voters int    `json:"voters"`
	Chosen bool   `json:"chosen"`
}

// Contact represents contact information
type Contact struct {
	PhoneNumber string `json:"phone_number"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	UserID      int64  `json:"user_id"`
}

// Location represents location information
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
