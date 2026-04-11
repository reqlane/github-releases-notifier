package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

type SubscriptionRepository struct {
	db *sql.DB
}

func NewSubcriptionRepository(db *sql.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
	query := `SELECT s.email, r.repo, s.confirmed, r.last_seen_tag FROM subscriptions s JOIN repos r ON s.repo_id = r.id WHERE email=?`
	rows, err := r.db.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscriptionsByEmail: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)
	for rows.Next() {
		var subscription model.Subscription
		err = rows.Scan(&subscription.Email, &subscription.Repo, &subscription.Confirmed, &subscription.LastSeenTag)
		if err != nil {
			return nil, fmt.Errorf("repository.GetSubscriptionsByEmail: %w", err)
		}
		subscriptions = append(subscriptions, subscription)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscriptionsByEmail: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) CreateSubscription(email string, repoID int, confirmToken, unsubscribeToken string) error {
	query := `INSERT INTO subscriptions (email, repo_id, confirm_token, unsubscribe_token) VALUES (?,?,?,?)`
	_, err := r.db.Exec(query, email, repoID, confirmToken, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("repository.CreateSubscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) SubscriptionExists(email string, repoName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions s JOIN repos r ON s.repo_id = r.id WHERE s.email=? AND r.repo=?)`
	var exists bool
	err := r.db.QueryRow(query, email, repoName).Scan(exists)
	if err != nil {
		return false, fmt.Errorf("repository.SubscriptionExists: %w", err)
	}
	return exists, nil
}

func (r *SubscriptionRepository) GetRepoByName(repoName string) (model.Repo, error) {
	query := `SELECT id, repo, last_seen_tag FROM repos WHERE repo=?`
	var repo model.Repo

	err := r.db.QueryRow(query, repoName).Scan(&repo.ID, &repo.Repo, &repo.LastSeenTag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Repo{}, apperror.ErrNotFound
		}
		return model.Repo{}, fmt.Errorf("repository.GetRepoByName: %w", err)
	}

	return repo, nil
}

func (r *SubscriptionRepository) CreateRepo(repo model.Repo) (model.Repo, error) {
	query := `INSERT INTO repos (repo, last_seen_tag) VALUES (?,?)`

	result, err := r.db.Exec(query, repo.Repo, repo.LastSeenTag)
	if err != nil {
		if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
			if mysqlErr.Number == 1062 {
				return model.Repo{}, apperror.ErrAlreadyExists
			}
		}
		return model.Repo{}, fmt.Errorf("repository.CreateRepo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return model.Repo{}, fmt.Errorf("repository.CreateRepo: %w", err)
	}

	repo.ID = int(id)
	return repo, nil
}
