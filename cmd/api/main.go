package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
	hub := ws.NewHub(convService)
	go hub.Run()
	wsHander := ws.NewWSHanler(hub, convService)

	// router
	handlers := router.Handlers{
		User: userHandler,
		Auth: authHandler,
		Conv: convHandler,
		WS:   wsHander,
	}

	router := router.NewRouter(handlers, ctx)

	serverAddr := config.ENV.GetServerAddr()

	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,  // Defends against Slowloris
		ReadTimeout:       10 * time.Second, // Max time to read request
		WriteTimeout:      30 * time.Second, // Max time to write response
		IdleTimeout:       60 * time.Second, // Keep-alive idle limit
		MaxHeaderBytes:    1 << 20,          // 1 MB
	}

	// Run server in background goroutine
	go func() {
		log.Printf("Gin API server listening on Address %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error %v", err)
		}
	}()

	// Wait for OS interrupt signals
	<-ctx.Done()
	log.Printf("Received shutdown signal, initiating graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	// Drain background conversation wipeout workers
	if err := convService.Shutdown(shutdownCtx); err != nil {
		log.Printf("Wipeout workers drain timed out: %v", err)
	}

	log.Println("Server shutdown complete.")
}
