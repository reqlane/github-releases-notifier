package repository

import "database/sql"

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubcriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}
