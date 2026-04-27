package app

import (
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/api/router"
	"github.com/reqlane/github-releases-notifier/internal/config"
	"github.com/reqlane/github-releases-notifier/internal/contract"
	"github.com/reqlane/github-releases-notifier/internal/db"
	"github.com/reqlane/github-releases-notifier/internal/githubapi"
	"github.com/reqlane/github-releases-notifier/internal/notifier/gomail"
	"github.com/reqlane/github-releases-notifier/internal/repository/mariadb"
	"github.com/reqlane/github-releases-notifier/internal/scanner"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/rs/zerolog"
)

func Run() (err error) {
	logger := zerolog.New(os.Stderr).
		With().
		Timestamp().
		Logger()

	err = godotenv.Load(".env")
	if err != nil {
		logger.Warn().Str("error message", err.Error()).Msg("error loading .env file, assuming docker is being used")
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Err(err).Msg("error loading config")
		return
	}

	dbConn, err := db.ConnectDB(cfg.DSN())
	if err != nil {
		logger.Err(err).Msg("error connecting to db")
		return
	}
	defer func() {
		sqlDB, err := dbConn.DB()
		if err != nil {
			logger.Err(err).Msg("error getting sql.DB from gorm")
			return
		}
		if err = sqlDB.Close(); err != nil {
			logger.Err(err).Msg("error closing db connection")
		}
	}()

	err = db.RunMigrations(dbConn)
	if err != nil {
		logger.Err(err).Msg("error running migrations")
		return
	}

	httpclient := http.Client{Timeout: 10 * time.Second}
	githubClient := githubapi.NewClient(&httpclient, logger, cfg.GithubToken)

	notifier := initNotifier(cfg)

	repository := mariadb.NewSubscriptionRepo(dbConn)
	subscriptionUseCase := usecase.NewSubscriptionUseCase(repository, githubClient, notifier)
	subscriptionHandler := handler.NewSubcriptionHandler(subscriptionUseCase, logger)

	scan := initScanner(repository, githubClient, notifier, logger, cfg.GithubToken)
	go scan.Run()

	rt := router.NewRouter(subscriptionHandler)
	engine := rt.Build()

	err = engine.Run(":" + cfg.ServerPort)
	if err != nil {
		return
	}

	return
}

func initNotifier(cfg *config.Config) contract.Notifier {
	return gomail.NewNotifier(gomail.GomailNotifierConfig{
		Host:          cfg.SMTPHost,
		Port:          cfg.SMTPPort,
		Username:      cfg.SMTPUsername,
		Password:      cfg.SMTPPassword,
		ServerBaseURL: cfg.ServerBaseURL,
	})
}

func initScanner(r contract.SubscriptionRepo, g contract.GithubClient, n contract.Notifier, l zerolog.Logger, githubAPIToken string) *scanner.FixedRateScanner {
	scan := scanner.NewFixedRateScanner(r, g, n, l)
	if githubAPIToken == "" {
		scan.SetRequestsPerMin(1)
	}
	return scan
}
