package mariadb

import (
	"errors"

	"github.com/reqlane/github-releases-notifier/internal/apperror"
	"github.com/reqlane/github-releases-notifier/internal/contract"
	"github.com/reqlane/github-releases-notifier/internal/entity"
	"github.com/reqlane/github-releases-notifier/internal/model"
	"gorm.io/gorm"
)

type mariadbSubscriptionRepo struct {
	db *gorm.DB
}

func NewSubscriptionRepo(db *gorm.DB) contract.SubscriptionRepo {
	return &mariadbSubscriptionRepo{db: db}
}

func (r *mariadbSubscriptionRepo) GetSubscriptionsByEmail(email string) ([]model.Subscription, error) {
	subscriptions := make([]model.Subscription, 0)
	result := r.db.
		Select(`subscriptions.email, Repo.repo, subscriptions.confirmed, Repo.last_seen_tag`).
		Model(&entity.Subscription{}).
		InnerJoins("Repo").
		Where(`subscriptions.email = ?`, email).
		Scan(&subscriptions)
	if result.Error != nil {
		return nil, result.Error
	}
	return subscriptions, nil
}

func (r *mariadbSubscriptionRepo) CreateSubscription(email string, repoID uint, confirmToken, unsubscribeToken string) error {
	subscription := entity.Subscription{
		Email:            email,
		RepoID:           repoID,
		ConfirmToken:     &confirmToken,
		UnsubscribeToken: unsubscribeToken,
	}
	return r.db.Create(&subscription).Error
}

func (r *mariadbSubscriptionRepo) SubscriptionExists(email string, repoName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM subscriptions s JOIN repos r ON s.repo_id = r.id WHERE s.email=? AND r.repo=?)`
	var exists bool
	result := r.db.
		Raw(query, email, repoName).
		Scan(&exists)
	return exists, result.Error
}

func (r *mariadbSubscriptionRepo) ConfirmSubscription(confirmToken string) error {
	result := r.db.
		Model(&entity.Subscription{}).
		Where(`confirm_token = ?`, confirmToken).
		Updates(map[string]any{"confirmed": true, "confirm_token": nil})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *mariadbSubscriptionRepo) DeleteSubscription(unsubscribeToken string) error {
	result := r.db.
		Where(`unsubscribe_token = ?`, unsubscribeToken).
		Delete(&entity.Subscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *mariadbSubscriptionRepo) GetRepoByName(repoName string) (model.Repo, error) {
	var repo entity.Repo
	result := r.db.
		Where(`repo = ?`, repoName).
		First(&repo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return model.Repo{}, apperror.ErrNotFound
		}
		return model.Repo{}, result.Error
	}
	return model.Repo{ID: repo.ID, Repo: repo.Repo, LastSeenTag: repo.LastSeenTag}, nil
}

func (r *mariadbSubscriptionRepo) CreateRepo(repoName, lastSeenTag string) (model.Repo, error) {
	repo := entity.Repo{Repo: repoName, LastSeenTag: lastSeenTag}
	result := r.db.Create(&repo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return model.Repo{}, apperror.ErrAlreadyExists
		}
		return model.Repo{}, result.Error
	}
	return model.Repo{ID: repo.ID, Repo: repo.Repo, LastSeenTag: repo.LastSeenTag}, nil
}

func (r *mariadbSubscriptionRepo) GetSubscribedRepos() ([]model.Repo, error) {
	repos := make([]entity.Repo, 0)
	result := r.db.
		Where(`EXISTS (SELECT 1 FROM subscriptions WHERE repo_id=repos.id AND confirmed=true)`).
		Find(&repos)
	if result.Error != nil {
		return nil, result.Error
	}

	models := make([]model.Repo, len(repos))
	for i, e := range repos {
		models[i] = model.Repo{ID: e.ID, Repo: e.Repo, LastSeenTag: e.LastSeenTag}
	}
	return models, nil
}

func (r *mariadbSubscriptionRepo) GetNotificationTargetsByRepo(repoID uint) ([]model.NotificationTarget, error) {
	targets := make([]model.NotificationTarget, 0)
	result := r.db.
		Select(`email, unsubscribe_token`).
		Model(&entity.Subscription{}).
		Where(`repo_id = ? AND confirmed = true`, repoID).
		Scan(&targets)
	if result.Error != nil {
		return nil, result.Error
	}
	return targets, nil
}

func (r *mariadbSubscriptionRepo) UpdateLastSeenTag(repoID uint, tag string) error {
	return r.db.
		Model(&entity.Repo{ID: repoID}).
		Update("last_seen_tag", tag).
		Error
}
