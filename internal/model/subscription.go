package model

type Subscription struct {
	Email       string `json:"email" validate:"required,email"`
	Repo        string `json:"repo" validate:"required,github_repo"`
	Confirmed   bool   `json:"confirmed"`
	LastSeenTag string `json:"last_seen_tag"`
}

type Repo struct {
	ID          int
	Repo        string
	LastSeenTag string
}

type NotificationTarget struct {
	Email            string
	UnsubscribeToken string
}
