package server

import (
	"bank_system/pkg/account"
	"bank_system/pkg/transaction"
	"bank_system/pkg/user"
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Server struct {
	logger        *log.Logger
	pool          *pgxpool.Pool
	router        *gin.Engine
	actController *account.AccountController
	usrController *user.UserController
	txController  *transaction.TxController
	cron          *CronService
}

func NewServer() (*Server, error) {
	logger := log.Default()

	pool, err := SetPGConn(context.Background(), viper.GetString("postgres.connection_string"))
	if err != nil {
		fmt.Println("Failed to create connection pool", zap.Error(err))
		return nil, err
	}

	usrRepo := user.NewUserRepository(pool)
	usrService := user.NewUserService(usrRepo)
	usrController := user.NewUserController(usrService, logger)

	txRepo := transaction.NewTxRepository(pool)
	txService := transaction.NewTxService(txRepo)
	txController := transaction.NewTxController(txService, logger)

	actRepo := account.NewAccountRepository(pool)
	actService := account.NewAccountService(actRepo)
	actController := account.NewAccountController(actService, logger)

	cronService, err := NewCronService(pool, logger)
	if err != nil {
		return nil, err
	}

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	usrController.RegisterRoutes(router)
	txController.RegisterRoutes(router)
	actController.RegisterRoutes(router)

	return &Server{
		logger:        logger,
		pool:          pool,
		router:        router,
		actController: actController,
		usrController: usrController,
		txController:  txController,
		cron:          cronService,
	}, nil
}

func (s *Server) Start(port, certFile, keyFile string) error {
	err := s.cron.Start()
	if err != nil {
		s.logger.Printf("Failed to start cronjob: %v\n", err)
		return err
	}

	return s.router.RunTLS(":"+port, certFile, keyFile)
}

func (s *Server) Stop() {
	s.pool.Close()
	s.cron.Stop()
}
