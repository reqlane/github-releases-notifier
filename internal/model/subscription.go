package model

type Subscription struct {
	Email       string `json:"email"`
	Repo        string `json:"repo"`
	Confirmed   bool   `json:"confirmed"`
	LastSeenTag string `json:"last_seen_tag"`
}

type SubscribeRequest struct {
	Email string `json:"email" validate:"required,email"`
	Repo  string `json:"repo" validate:"required,github_repo"`
}

type SubscriptionFilter struct {
	Email string `json:"email" validate:"required,email"`
}
