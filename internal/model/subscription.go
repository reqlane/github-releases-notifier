package model

type Subscription struct {
	email       string
	repo        string
	confirmed   bool
	lastSeenTag string
}

type SubscribeRequest struct {
	email string
	repo  string
}

type UnsubscribeResponseData struct {
	Email             string   `json:"email"`
	UnsubscribedCount int      `json:"unsubscribed_count"`
	UnsubscribedRepos []string `json:"unsubscribed_repos"`
}
