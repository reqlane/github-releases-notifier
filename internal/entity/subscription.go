package entity

type Subscription struct {
	ID               uint `gorm:"primaryKey"`
	Email            string
	RepoID           uint
	Confirmed        bool
	ConfirmToken     *string
	UnsubscribeToken string
	Repo             Repo `gorm:"foreignKey:RepoID"`
}
