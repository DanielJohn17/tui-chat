package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/DanielJohn17/tui-chat/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/internal/api/config"
	"github.com/DanielJohn17/tui-chat/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/internal/api/database"
	"github.com/DanielJohn17/tui-chat/internal/api/db"
	"github.com/DanielJohn17/tui-chat/internal/api/router"
	"github.com/DanielJohn17/tui-chat/internal/api/users"
	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbURL := config.ENV.GetDBURL()

	ctx := context.Background()

	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	sqlDB := stdlib.OpenDBFromPool(conn)
	defer func() {
		_ = sqlDB.Close()
		conn.Close()
	}()

	if err := db.Migrate(sqlDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	q := database.New(conn)

	// users
	userRepo := users.NewUserRepository(q)
	userService := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userService)

	// auth
	authService := auth.NewAuthService(userService)
	authHandler := auth.NewAuthHander(authService)

	// conversations
	convRepo := conversations.NewConvRepository(q)
	convService := conversations.NewConvService(convRepo, userService)
	convHandler := conversations.NewConvHandler(convService)

	// websocket conn for conversations
	hub := ws.NewHub()
	go hub.Run()
	wsHander := ws.NewWSHanler(hub, convService)

	// router
	handlers := router.Handlers{
		User: userHandler,
		Auth: authHandler,
		Conv: convHandler,
		WS:   wsHander,
	}

	router := router.NewRouter(handlers)

	serverAddr := config.ENV.GetServerAddr()
	if err := router.Run(serverAddr); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start gin server: %v\n", err)
		os.Exit(1)
	}
}
