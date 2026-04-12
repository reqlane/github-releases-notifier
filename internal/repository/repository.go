package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/model"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
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

func (r *Repository) CreateSubscription(email string, repoID int, confirmToken, unsubscribeToken string) error {
	query := `INSERT INTO subscriptions (email, repo_id, confirm_token, unsubscribe_token) VALUES (?,?,?,?)`
	_, err := r.db.Exec(query, email, repoID, confirmToken, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("repository.CreateSubscription: %w", err)
	}
	return nil
}

func (r *Repository) SubscriptionExists(email string, repoName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions s JOIN repos r ON s.repo_id = r.id WHERE s.email=? AND r.repo=?)`
	var exists bool
	err := r.db.QueryRow(query, email, repoName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("repository.SubscriptionExists: %w", err)
	}
	return exists, nil
}

func (r *Repository) ConfirmSubscription(confirmToken string) error {
	query := `UPDATE subscriptions SET confirmed=true, confirm_token=NULL WHERE confirm_token=?`
	res, err := r.db.Exec(query, confirmToken)
	if err != nil {
		return fmt.Errorf("repository.ConfirmSubscription: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.ConfirmSubscription: %w", err)
	}
	if rowsAffected == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *Repository) DeleteSubscription(unsubscribeToken string) error {
	query := `DELETE FROM subscriptions WHERE unsubscribe_token=?`
	res, err := r.db.Exec(query, unsubscribeToken)
	if err != nil {
		return fmt.Errorf("repository.DeleteSubscription: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("repository.DeleteSubscription: %w", err)
	}
	if rowsAffected == 0 {
		return apperror.ErrNotFound
	}

	return nil
}

func (r *Repository) GetRepoByName(repoName string) (model.Repo, error) {
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

func (r *Repository) CreateRepo(repo model.Repo) (model.Repo, error) {
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

func (r *Repository) GetSubscribedRepos() ([]model.Repo, error) {
	query := `SELECT id, repo, last_seen_tag FROM repos WHERE EXISTS (SELECT 1 FROM subscriptions WHERE repo_id=repos.id AND confirmed=true)`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscribedRepos: %w", err)
	}
	defer rows.Close()

	repos := make([]model.Repo, 0)
	for rows.Next() {
		var repo model.Repo
		err = rows.Scan(&repo.ID, &repo.Repo, &repo.LastSeenTag)
		if err != nil {
			return nil, fmt.Errorf("repository.GetSubscribedRepos: %w", err)
		}
		repos = append(repos, repo)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscribedRepos: %w", err)
	}

	return repos, nil
}

func (r *Repository) GetNotificationTargetsByRepo(repoID int) ([]model.NotificationTarget, error) {
	query := `SELECT email, unsubscribe_token FROM subscribers WHERE repo_id=? AND confirmed=true`
	rows, err := r.db.Query(query, repoID)
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscribersByRepo: %w", err)
	}
	defer rows.Close()

	targets := make([]model.NotificationTarget, 0)
	for rows.Next() {
		var target model.NotificationTarget
		err = rows.Scan(&target.Email, &target.UnsubscribeToken)
		if err != nil {
			return nil, fmt.Errorf("repository.GetSubscribersByRepo: %w", err)
		}
		targets = append(targets, target)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("repository.GetSubscribersByRepo: %w", err)
	}

	return targets, nil
}

func (r *Repository) UpdateLastSeenTag(repoID int, tag string) error {
	query := `UPDATE repos SET last_seen_tag=? WHERE id=?`
	_, err := r.db.Exec(query, tag, repoID)
	if err != nil {
		return fmt.Errorf("repository.UpdateLastSeenTag: %w", err)
	}
	return nil
}
