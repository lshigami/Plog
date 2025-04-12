package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lshigami/Plog/internal/api"
	"github.com/lshigami/Plog/internal/config"
	"github.com/lshigami/Plog/internal/db/sqlc"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	connPool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)

	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}
	defer connPool.Close()

	store := sqlc.New(connPool)

	router := api.SetupRouter(store, *cfg)

	serverAddress := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s\n", serverAddress)
	err = router.Run(serverAddress)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
