package slack

type Channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IsPrivate   bool   `json:"is_private"`
	IsArchived  bool   `json:"is_archived"`
	IsMember    bool   `json:"is_member"`
	NumMembers  int    `json:"num_members"`
	Topic       Topic  `json:"topic"`
	Purpose     Topic  `json:"purpose"`
	Created     int64  `json:"created"`
}

type Topic struct {
	Value   string `json:"value"`
	Creator string `json:"creator"`
	LastSet int64  `json:"last_set"`
}

type User struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	RealName  string  `json:"real_name"`
	Email     string  `json:"email"`
	IsBot     bool    `json:"is_bot"`
	IsAdmin   bool    `json:"is_admin"`
	IsOwner   bool    `json:"is_owner"`
	Deleted   bool    `json:"deleted"`
	Profile   Profile `json:"profile"`
}

type Profile struct {
	Email       string `json:"email"`
	RealName    string `json:"real_name"`
	DisplayName string `json:"display_name"`
	StatusText  string `json:"status_text"`
	StatusEmoji string `json:"status_emoji"`
	Image48     string `json:"image_48"`
	Image192    string `json:"image_192"`
}

type Message struct {
	Channel   string `json:"channel"`
	Text      string `json:"text"`
	Timestamp string `json:"ts"`
	User      string `json:"user"`
	Type      string `json:"type"`
}

type WebhookPayload struct {
	Text        string        `json:"text,omitempty"`
	Channel     string        `json:"channel,omitempty"`
	Username    string        `json:"username,omitempty"`
	IconEmoji   string        `json:"icon_emoji,omitempty"`
	IconURL     string        `json:"icon_url,omitempty"`
	Attachments []interface{} `json:"attachments,omitempty"`
	Blocks      []interface{} `json:"blocks,omitempty"`
}

type SendMessageRequest struct {
	Channel  string        `json:"channel"`
	Text     string        `json:"text,omitempty"`
	Blocks   []interface{} `json:"blocks,omitempty"`
	ThreadTS string        `json:"thread_ts,omitempty"`
	Mrkdwn   bool          `json:"mrkdwn,omitempty"`
}
