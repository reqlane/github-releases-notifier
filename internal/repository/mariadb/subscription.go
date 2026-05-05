package mariadb

import (
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
	var rows []entity.Subscription
	result := r.db.
		Model(&entity.Subscription{}).
		InnerJoins("Repo").
		Where(entity.Subscription{Email: email}).
		Scan(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	subscriptions := make([]model.Subscription, len(rows))
	for i, row := range rows {
		lastSeenTagValue := ""
		if row.Repo.LastSeenTag != nil {
			lastSeenTagValue = *row.Repo.LastSeenTag
		}
		subscriptions[i] = model.Subscription{
			Email:       row.Email,
			Repo:        row.Repo.Repo,
			Confirmed:   row.Confirmed,
			LastSeenTag: lastSeenTagValue,
		}
	}

	return subscriptions, nil
}

func (r *mariadbSubscriptionRepo) CreateSubscription(email string, repoID uint, tokens model.SubscriptionTokens) error {
	subscription := entity.Subscription{
		Email:            email,
		RepoID:           repoID,
		ConfirmToken:     &tokens.ConfirmToken,
		UnsubscribeToken: tokens.UnsubscribeToken,
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
		Where(entity.Subscription{ConfirmToken: &confirmToken}).
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
		Where(entity.Subscription{UnsubscribeToken: unsubscribeToken}).
		Delete(&entity.Subscription{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperror.ErrNotFound
	}
	return nil
}

func (r *mariadbSubscriptionRepo) GetOrCreateRepo(repoName string, lastSeenTag string) (model.Repo, error) {
	var repo entity.Repo
	result := r.db.
		Where(entity.Repo{Repo: repoName}).
		Attrs(entity.Repo{LastSeenTag: &lastSeenTag}).
		FirstOrCreate(&repo)
	if result.Error != nil {
		return model.Repo{}, result.Error
	}
	lastSeenTagValue := ""
	if repo.LastSeenTag != nil {
		lastSeenTagValue = *repo.LastSeenTag
	}
	return model.Repo{ID: repo.ID, Repo: repo.Repo, LastSeenTag: lastSeenTagValue}, nil
}

func (r *mariadbSubscriptionRepo) GetSubscribedRepos() ([]model.Repo, error) {
	var rows []entity.Repo
	result := r.db.
		Where(`EXISTS (SELECT 1 FROM subscriptions WHERE repo_id=repos.id AND confirmed=true)`).
		Find(&rows)
	if result.Error != nil {
		return nil, result.Error
	}

	repos := make([]model.Repo, len(rows))
	for i, e := range rows {
		lastSeenTagValue := ""
		if e.LastSeenTag != nil {
			lastSeenTagValue = *e.LastSeenTag
		}
		repos[i] = model.Repo{ID: e.ID, Repo: e.Repo, LastSeenTag: lastSeenTagValue}
	}
	return repos, nil
}

func (r *mariadbSubscriptionRepo) GetNotificationTargetsByRepo(repoID uint) ([]model.NotificationTarget, error) {
	targets := make([]model.NotificationTarget, 0)
	result := r.db.
		Select(`email, unsubscribe_token`).
		Model(&entity.Subscription{}).
		Where(entity.Subscription{RepoID: repoID, Confirmed: true}).
		Scan(&targets)
	if result.Error != nil {
		return nil, result.Error
	}
	return targets, nil
}

func (r *mariadbSubscriptionRepo) UpdateLastSeenTag(repoID uint, tag string) error {
	return r.db.
		Model(&entity.Repo{ID: repoID}).
		Update("last_seen_tag", &tag).
		Error
}
