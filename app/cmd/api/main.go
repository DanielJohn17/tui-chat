package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DanielJohn17/tui-chat/app/internal/api/auth"
	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/DanielJohn17/tui-chat/app/internal/api/conversations"
	"github.com/DanielJohn17/tui-chat/app/internal/api/database"
	"github.com/DanielJohn17/tui-chat/app/internal/api/router"
	"github.com/DanielJohn17/tui-chat/app/internal/api/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		config.ENV.DBUser,
		config.ENV.DBPassword,
		config.ENV.DBHost,
		config.ENV.DBport,
		config.ENV.DBName,
	)

	ctx := context.Background()

	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
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

	// router
	handlers := router.Handlers{
		User: userHandler,
		Auth: authHandler,
		Conv: convHandler,
	}

	router := router.NewRouter(handlers)

	if err := router.Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to start gin server: %v\n", err)
		os.Exit(1)
	}
}
