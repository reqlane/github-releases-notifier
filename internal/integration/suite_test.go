package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	mysqlmigrate "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/reqlane/github-releases-notifier/internal/api/handler"
	"github.com/reqlane/github-releases-notifier/internal/db"
	mockgithubapi "github.com/reqlane/github-releases-notifier/internal/mock/githubapi"
	mocknotifier "github.com/reqlane/github-releases-notifier/internal/mock/notifier"
	"github.com/reqlane/github-releases-notifier/internal/repository/mariadb"
	"github.com/reqlane/github-releases-notifier/internal/usecase"
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
	repo := mariadb.NewSubscriptionRepo(s.db)
	uc := usecase.NewSubscriptionUseCase(repo, s.ghclient, s.notif)
	h := handler.NewSubcriptionHandler(uc, zerolog.New(io.Discard))
	s.router = setupRouter(h)
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

func (s *IntegrationTestSuite) TestTemplate() {
	fmt.Println("Just for testing CI")
}
