package server

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"bank_system/pkg/account"
	"bank_system/pkg/transaction"
	"bank_system/pkg/user"

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CronService struct {
	scheduler  gocron.Scheduler
	logger     *log.Logger
	usrService *user.UserService
	actService *account.AccountService
	txService  *transaction.TxService
}

func NewCronService(pool *pgxpool.Pool, logger *log.Logger) (*CronService, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	usrRepo := user.NewUserRepository(pool)
	usrService := user.NewUserService(usrRepo)

	txRepo := transaction.NewTxRepository(pool)
	txService := transaction.NewTxService(txRepo)

	actRepo := account.NewAccountRepository(pool)
	actService := account.NewAccountService(actRepo)

	return &CronService{
		scheduler:  s,
		logger:     logger,
		usrService: usrService,
		actService: actService,
		txService:  txService,
	}, nil
}

func (c *CronService) Start() error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Job: Create user, account, and transaction
	_, err := c.scheduler.NewJob(
		gocron.DurationJob(
			1*time.Minute,
		),
		gocron.NewTask(
			func(logger *log.Logger) {
				rInt := r.Uint32()

				username := fmt.Sprintf("user_%d", rInt)
				email := username + "@example.com"
				password := fmt.Sprintf("password_%d", rInt)

				ctx := context.Background()

				user, err := c.usrService.CreateUser(ctx, username, email, password)
				if err != nil {
					logger.Printf("cronjob 1 - create user failed: %v\n", err)
					return
				}

				account, err := c.actService.CreateAccount(ctx, user.ID)
				if err != nil {
					logger.Printf("cronjob 1 - create account failed: %v\n", err)
					return
				}

				txId, balance, err := c.actService.Deposit(ctx, account.ID, float64(rInt), "")
				if err != nil {
					logger.Printf("cronjob 1 - create transaction failed: %v\n", err)
					return
				}

				logger.Printf(
					"cronjob 1 - deposited %f from account %d (tx_id: %d), new balance: %f\n",
					float64(rInt), account.ID, txId, balance,
				)
			},
			c.logger,
		),
	)

	if err != nil {
		return err
	}

	// Job: Create account, and transaction
	_, err = c.scheduler.NewJob(
		gocron.DurationJob(
			12*time.Hour,
		),
		gocron.NewTask(
			func(logger *log.Logger) {
				rInt := r.Uint32()

				users, err := c.usrService.GetAllUsers(context.Background())
				if err != nil {
					logger.Printf("cronjob 2 - get all users failed: %v\n", err)
					return
				}

				for _, user := range *users {
					ctx := context.Background()
					account, err := c.actService.CreateAccount(ctx, user.ID)
					if err != nil {
						logger.Printf("cronjob 2 - create account failed: %v\n", err)
						return
					}

					txId, balance, err := c.actService.Deposit(ctx, account.ID, float64(rInt), "")
					if err != nil {
						logger.Printf("cronjob 2 - create transaction failed: %v\n", err)
						return
					}

					logger.Printf(
						"cronjob 2 - deposited %f to account %d (tx_id: %d), new balance: %f\n",
						float64(rInt), account.ID, txId, balance,
					)
				}
			},
			c.logger,
		),
	)

	if err != nil {
		return err
	}

	// Job: Create transaction
	_, err = c.scheduler.NewJob(
		gocron.DurationJob(
			30*time.Second,
		),
		gocron.NewTask(
			func(logger *log.Logger) {
				rInt := r.Intn(1000000)

				accounts, err := c.actService.GetAllAccounts(context.Background())
				if err != nil {
					logger.Printf("cronjob 3 - get all accounts failed: %v\n", err)
					return
				}

				for _, account := range accounts {
					txId, balance, err := c.actService.Withdraw(context.Background(), account.IDNumber, float64(rInt), "")
					if err != nil {
						logger.Printf("cronjob 3 - create transaction failed: %v\n", err)
						return
					}

					logger.Printf(
						"cronjob 3 - withdraw %f from account %d (tx_id: %d), new balance: %f\n",
						float64(rInt), account.ID, txId, balance,
					)
				}
			},
			c.logger,
		),
	)

	if err != nil {
		return err
	}

	c.scheduler.Start()
	c.logger.Printf("Cron jobs started successfully\n")

	return nil
}

func (c *CronService) Stop() error {
	err := c.scheduler.Shutdown()
	if err != nil {
		c.logger.Printf("Failed to stop cron jobs: %v\n", err)
		return err
	}
	c.logger.Printf("Cron jobs stopped successfully\n")
	return nil
}

// func getModel(i uint32) *simulator.LLModel {
// 	depositModel := simulator.NewLLModel(
// 		[]simulator.Message{
// 			{
// 				Role:    viper.GetString("openrouter.role.user"),
// 				Content: viper.GetString("openrouter.message.deposit"),
// 			},
// 		})

// 	withdrawModel := simulator.NewLLModel(
// 		[]simulator.Message{
// 			{
// 				Role:    viper.GetString("openrouter.role.user"),
// 				Content: viper.GetString("openrouter.message.withdraw"),
// 			},
// 		})

// 	if i%2 == 0 {
// 		return depositModel
// 	}
// 	return withdrawModel
// }
