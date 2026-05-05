package usecase

type SubscribeInput struct {
	Email string `label:"email" validate:"required,email"`
	Repo  string `label:"repo" validate:"required,github_repo"`
}
