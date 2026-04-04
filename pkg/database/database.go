package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustInitDB() *pgxpool.Pool {
	var dbPool *pgxpool.Pool

	config, err := pgxpool.ParseConfig(getConnectString())
	if err != nil {
		log.Fatalf("Unable parse connection string: %v\n", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	pingAttemptsLimit := 3
	var pingErr error

	for i := 1; i <= pingAttemptsLimit; i++ {
		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr = dbPool.Ping(pingCtx)
		pingCancel()

		if pingErr == nil {
			break
		}
		log.Printf("db ping attempt %d failed: %v", i, pingErr)
		if i < pingAttemptsLimit {
			time.Sleep(300 * time.Millisecond)
		}
	}

	if pingErr != nil {
		log.Fatalf("Unable ti ping databse")
	}

	log.Println("Database connection pool established")
	return dbPool
}

func getConnectString() string {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, db)
}
