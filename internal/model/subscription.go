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
