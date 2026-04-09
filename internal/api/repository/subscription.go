package repository

import (
	"database/sql"
	"fmt"

	"github.com/reqlane/github-releases-notifier/internal/model"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubcriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetByEmail(email string) ([]model.Subscription, error) {
	query := `SELECT email, repo, confirmed, last_seen_tag FROM subscriptions WHERE email=?`
	rows, err := r.db.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("repository.GetByEmail: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)
	for rows.Next() {
		var subscription model.Subscription
		err = rows.Scan(&subscription.Email, &subscription.Repo, &subscription.Confirmed, &subscription.LastSeenTag)
		if err != nil {
			return nil, fmt.Errorf("repository.GetByEmail: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repository.GetByEmail: %w", err)
	}

	return subscriptions, nil
}
