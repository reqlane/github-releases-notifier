package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/db"
	"github.com/reqlane/github-releases-notifier/internal/infrastructure/repository/mariadb"
	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
	"github.com/reqlane/github-releases-notifier/pkg/tokengen"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	tcmariadb "github.com/testcontainers/testcontainers-go/modules/mariadb"
	"gorm.io/gorm"
)

type IntegrationTestSuite struct {
	suite.Suite

	ctx       context.Context
	container *tcmariadb.MariaDBContainer
	db        *gorm.DB
	router    *gin.Engine
	ghclient  *mockgithubapi.GithubClient
	notif     *mocknotifier.Notifier
	tokenGen  tokengen.Generator
}

func TestIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := tcmariadb.Run(s.ctx,
		"mariadb:11",
		tcmariadb.WithDatabase("integration_test_db"),
		tcmariadb.WithUsername("test"),
		tcmariadb.WithPassword("test"),
	)
	s.Require().NoError(err)
	s.container = container

	dsn, err := container.ConnectionString(s.ctx, "parseTime=True")
	s.Require().NoError(err)

	gormdb, err := db.ConnectDB(dsn)
	s.Require().NoError(err)
	s.db = gormdb

	s.runMigrations("../../migrations")
	s.initDependencies()
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.Require().NoError(s.container.Terminate(s.ctx))
}

func (s *IntegrationTestSuite) SetupTest() {
	cleanupQueries := []string{
		"SET FOREIGN_KEY_CHECKS = 0",
		"TRUNCATE TABLE subscriptions",
		"TRUNCATE TABLE repos",
		"SET FOREIGN_KEY_CHECKS = 1",
	}
	for _, query := range cleanupQueries {
		s.Require().NoError(s.db.Exec(query).Error)
	}
	s.ghclient.ExpectedCalls = nil
	s.ghclient.Calls = nil
	s.notif.ExpectedCalls = nil
	s.notif.Calls = nil
}

func (s *IntegrationTestSuite) runMigrations(migrationsPath string) {
	sqlDB, err := s.db.DB()
	s.Require().NoError(err)

	driver, err := mysqlmigrate.WithInstance(sqlDB, &mysqlmigrate.Config{})
	s.Require().NoError(err)

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsPath), "mysql", driver)
	s.Require().NoError(err)

	err = m.Up()
	s.Require().True(err == nil || errors.Is(err, migrate.ErrNoChange))
}

func (s *IntegrationTestSuite) initDependencies() {
	s.ghclient = new(mockgithubapi.GithubClient)
	s.notif = new(mocknotifier.Notifier)
	s.tokenGen = tokengen.NewRandGenerator(32)
	repo := mariadb.NewSubscriptionRepo(s.db)
	uc := usecase.NewSubscriptionUseCase(repo, s.ghclient, s.notif)
	h := handler.NewSubcriptionHandler(uc, zerolog.New(io.Discard))
	s.router = setupRouter(h)
}

func (s *IntegrationTestSuite) performRequest(method, path string, body any) *httptest.ResponseRecorder {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		s.Require().NoError(err)
		reqBody = bytes.NewBuffer(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	return w
}

func (s *IntegrationTestSuite) seedSubscription(email, repo string, confirmed bool) (string, string) {
	s.Require().NoError(s.db.Exec(
		`INSERT INTO repos (repo, last_seen_tag) VALUES (?, ?) ON DUPLICATE KEY UPDATE id=id`,
		repo, "v1.0.0",
	).Error)

	var repoID uint
	s.Require().NoError(s.db.Raw(`SELECT id FROM repos WHERE repo=?`, repo).Scan(&repoID).Error)

	confirmToken := s.tokenGen.Generate()
	unsubscribeToken := s.tokenGen.Generate()
	var ct *string
	if !confirmed {
		ct = &confirmToken
	}
	s.Require().NoError(s.db.Exec(
		`INSERT INTO subscriptions (email, repo_id, confirmed, confirm_token, unsubscribe_token)
         VALUES (?, ?, ?, ?, ?)`,
		email, repoID, confirmed, ct, unsubscribeToken,
	).Error)

	return confirmToken, unsubscribeToken
}

func setupRouter(h *handler.SubscriptionHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	rt := gin.New()
	api := rt.Group("/api")
	api.POST("/subscribe", h.SubscribeHandler)
	api.GET("/confirm/:token", h.ConfirmHandler)
	api.GET("/unsubscribe/:token", h.UnsubscribeHandler)
	api.GET("/subscriptions", h.GetSubscriptionsHandler)
	return rt
}
