package usecase

type SubscribeInput struct {
	Email string `validate:"required,email"`
	Repo  string `validate:"required,github_repo"`
}
