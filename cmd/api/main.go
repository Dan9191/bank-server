package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/bank-service/internal/handlers"
	"github.com/bank-service/internal/middleware"
	"github.com/bank-service/internal/repositories"
	"github.com/bank-service/internal/services"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

const (
	dbHost     = "localhost"
	dbPort     = 5436
	dbUser     = "test"
	dbPassword = "test"
	dbName     = "test"
	jwtSecret  = "your_jwt_secret"
)

func main() {
	// Инициализация логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	// Подключение к базе данных с указанием search_path
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=bank",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	logger.Debug("Connecting to database with connection string: ", connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	// Проверка соединения
	logger.Debug("Pinging database")
	if err := db.Ping(); err != nil {
		logger.Fatal("Failed to ping database: ", err)
	}
	logger.Info("Database connection established")

	// Выполнение миграций
	logger.Debug("Running migrations")
	if err := runMigrations(db, logger); err != nil {
		logger.Fatal("Failed to run migrations: ", err)
	}

	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo, jwtSecret)
	accountService := services.NewAccountService(accountRepo, userRepo, transactionRepo, db)

	// Инициализация обработчиков
	userHandler := handlers.NewUserHandler(userService, logger)
	accountHandler := handlers.NewAccountHandler(accountService, logger)

	// Создание маршрутизатора
	router := mux.NewRouter()

	// Публичные эндпоинты
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods("GET")
	router.HandleFunc("/register", userHandler.Register).Methods("POST")
	router.HandleFunc("/login", userHandler.Login).Methods("POST")

	// Защищенные эндпоинты
	protected := router.PathPrefix("/").Subrouter()
	protected.Use(middleware.AuthMiddleware(jwtSecret, logger))
	protected.HandleFunc("/profile", userHandler.Profile).Methods("GET")
	protected.HandleFunc("/accounts", accountHandler.CreateAccount).Methods("POST")
	protected.HandleFunc("/accounts", accountHandler.GetAccounts).Methods("GET")
	protected.HandleFunc("/accounts/{id}/deposit", accountHandler.Deposit).Methods("POST")
	protected.HandleFunc("/accounts/{id}/withdraw", accountHandler.Withdraw).Methods("POST")
	protected.HandleFunc("/transfer", accountHandler.Transfer).Methods("POST")

	// Настройка сервера
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	logger.Info("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed to start: ", err)
	}
}

func runMigrations(db *sql.DB, logger *logrus.Logger) error {
	logger.Debug("Creating schema bank")
	_, err := db.Exec("CREATE SCHEMA IF NOT EXISTS bank")
	if err != nil {
		return fmt.Errorf("failed to create schema bank: %w", err)
	}

	logger.Debug("Enabling pgcrypto extension")
	_, err = db.Exec("CREATE EXTENSION IF NOT EXISTS pgcrypto")
	if err != nil {
		return fmt.Errorf("failed to enable pgcrypto: %w", err)
	}

	logger.Debug("Creating table bank.users")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.users table: %w", err)
	}

	logger.Debug("Creating table bank.accounts")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.accounts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES bank.users(id) ON DELETE CASCADE,
			balance NUMERIC(15, 2) DEFAULT 0.0,
			currency VARCHAR(3) DEFAULT 'RUB',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.accounts table: %w", err)
	}

	logger.Debug("Creating table bank.cards")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.cards (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT REFERENCES bank.accounts(id) ON DELETE CASCADE,
			card_number TEXT NOT NULL,
			expiry_date TEXT NOT NULL,
			cvv TEXT NOT NULL,
			hmac TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.cards table: %w", err)
	}

	logger.Debug("Creating table bank.transactions")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.transactions (
			id BIGSERIAL PRIMARY KEY,
			account_id BIGINT REFERENCES bank.accounts(id) ON DELETE CASCADE,
			amount NUMERIC(15, 2) NOT NULL,
			type VARCHAR(50) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.transactions table: %w", err)
	}

	logger.Debug("Creating table bank.credits")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.credits (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT REFERENCES bank.users(id) ON DELETE CASCADE,
			amount NUMERIC(15, 2) NOT NULL,
			interest_rate NUMERIC(5, 2) NOT NULL,
			term_months INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.credits table: %w", err)
	}

	logger.Debug("Creating table bank.payment_schedules")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bank.payment_schedules (
			id BIGSERIAL PRIMARY KEY,
			credit_id BIGINT REFERENCES bank.credits(id) ON DELETE CASCADE,
			payment_date DATE NOT NULL,
			amount NUMERIC(15, 2) NOT NULL,
			paid BOOLEAN DEFAULT FALSE,
			penalty NUMERIC(15, 2) DEFAULT 0.0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)`)
	if err != nil {
		return fmt.Errorf("failed to create bank.payment_schedules table: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
