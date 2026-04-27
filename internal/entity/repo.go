package entity

type Repo struct {
	ID          uint `gorm:"primaryKey"`
	Repo        string
	LastSeenTag string
}
