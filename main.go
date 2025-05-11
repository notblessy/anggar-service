package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/notblessy/anggar-service/db"
	"github.com/notblessy/anggar-service/repository"
	"github.com/notblessy/anggar-service/router"
	"github.com/notblessy/anggar-service/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("cannot load .env file")
	}

	postgres := db.NewPostgres()

	botToken := os.Getenv("TELEGRAM_API_TOKEN")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Failed to create bot:", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	openAi := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Path",
		},
	}))
	e.Use(middleware.CORS())
	e.Validator = &utils.Ghost{Validator: validator.New()}

	userRepo := repository.NewUserRepository(postgres)
	walletRepo := repository.NewWalletRepository(postgres)
	budgetRepo := repository.NewScopeRepository(postgres)
	transactionRepo := repository.NewTransactionRepository(postgres)
	openAiRepo := repository.NewHandler(openAi)
	capitalBotRepo := repository.NewCapitalBotRepository(postgres, bot, openAiRepo)

	httpService := router.NewHTTPService()
	httpService.RegisterPostgres(postgres)
	httpService.RegisterUserRepository(userRepo)
	httpService.RegisterWalletRepository(walletRepo)
	httpService.RegisterScopeRepository(budgetRepo)
	httpService.RegisterTransactionRepository(transactionRepo)

	httpService.Router(e)

	// Shared context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	// Bot listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Bot listener started")

		if err := capitalBotRepo.ListenMessage(ctx); err != nil {
			log.Printf("Bot listener error: %v", err)
		}
	}()

	// HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("HTTP server started")

		if err := e.Start(":3400"); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutdown signal received")

	// Initiate graceful shutdown
	cancel() // stop bot listener
	ctxTimeout, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := e.Shutdown(ctxTimeout); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	wg.Wait()
	log.Println("All services shut down gracefully")
}
