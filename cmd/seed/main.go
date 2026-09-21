package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/api/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	Name     string
	Username string
	Messages []seedMessage
}

type seedMessage struct {
	FromUser bool
	Content  string
	Offset   time.Duration
}

func main() {
	dbURL := config.ENV.GetDBURL()

	ctx := context.Background()
	conn, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("Password123"), 12)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// 1. Ensure or find danieljo17
	var danielID int64
	err = conn.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", "danieljo17").Scan(&danielID)
	if err != nil {
		err = conn.QueryRow(
			ctx,
			"INSERT INTO users (name, username, password) VALUES ($1, $2, $3) RETURNING id",
			"Daniel Johannes",
			"danieljo17",
			string(passwordHash),
		).Scan(&danielID)
		if err != nil {
			log.Fatalf("Failed to create danieljo17: %v", err)
		}
		fmt.Printf("Created user danieljo17 (ID: %d)\n", danielID)
	} else {
		fmt.Printf("Found existing user danieljo17 (ID: %d)\n", danielID)
	}

	// 2. Define users to seed with conversations
	usersToSeed := []seedUser{
		{
			Name:     "Alice Smith",
			Username: "alice",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hey Daniel! Have you tested the new TUI chat client?", Offset: 2 * time.Hour},
				{FromUser: false, Content: "Yes! The Bubble Tea UI and WebSocket integration are running smoothly.", Offset: 1*time.Hour + 50*time.Minute},
				{FromUser: true, Content: "Awesome! Let me know if you need help testing the Bloom filter component.", Offset: 1*time.Hour + 30*time.Minute},
				{FromUser: false, Content: "Will do, let's sync up on that tomorrow morning.", Offset: 1 * time.Hour},
				{FromUser: true, Content: "Sounds great! See you then 👍", Offset: 20 * time.Minute},
			},
		},
		{
			Name:     "Bob Jones",
			Username: "bob",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hello Daniel, did you check the latest database migration files?", Offset: 5 * time.Hour},
				{FromUser: false, Content: "Checked them, the last_read_message_id index looks solid.", Offset: 4 * time.Hour},
				{FromUser: true, Content: "Nice! We can now track unread counts in real-time.", Offset: 3 * time.Hour},
			},
		},
		{
			Name:     "Charlie Brown",
			Username: "charlie",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hi Daniel! Coffee break soon?", Offset: 10 * time.Hour},
				{FromUser: false, Content: "Sure thing! Meet at the lobby in 15 mins?", Offset: 9*time.Hour + 45*time.Minute},
				{FromUser: true, Content: "Perfect!", Offset: 9*time.Hour + 40*time.Minute},
			},
		},
		{
			Name:     "Diana Prince",
			Username: "diana",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hey Daniel! The CORS middleware is working cleanly in production.", Offset: 1 * time.Hour},
				{FromUser: false, Content: "Great to hear! The proxy trust warning is resolved too.", Offset: 45 * time.Minute},
				{FromUser: true, Content: "Everything is looking release-ready 🚀", Offset: 10 * time.Minute},
			},
		},
		{
			Name:     "Ethan Hunt",
			Username: "ethan",
			Messages: []seedMessage{
				{FromUser: true, Content: "Daniel, can you verify the WebSocket heartbeat ping/pong interval?", Offset: 3 * time.Hour},
				{FromUser: false, Content: "Verified, 54s ping with 60s read deadline keeps the connection alive stably.", Offset: 2 * time.Hour},
				{FromUser: true, Content: "Mission accomplished.", Offset: 1*time.Hour + 15*time.Minute},
			},
		},
		{
			Name:     "Fiona Gallagher",
			Username: "fiona",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hey! Just wanted to say the terminal UI colors and layout look fantastic!", Offset: 30 * time.Minute},
				{FromUser: false, Content: "Thanks Fiona! Lipgloss styling made a huge difference.", Offset: 15 * time.Minute},
			},
		},
		{
			Name:     "George Clark",
			Username: "george",
			Messages: []seedMessage{
				{FromUser: true, Content: "Hey Daniel, let's connect the Bloom filter query endpoints when you're ready.", Offset: 8 * time.Hour},
			},
		},
	}

	for _, u := range usersToSeed {
		var uid int64
		err = conn.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", u.Username).Scan(&uid)
		if err != nil {
			err = conn.QueryRow(
				ctx,
				"INSERT INTO users (name, username, password) VALUES ($1, $2, $3) RETURNING id",
				u.Name,
				u.Username,
				string(passwordHash),
			).Scan(&uid)
			if err != nil {
				log.Printf("Failed to create user %s: %v", u.Username, err)
				continue
			}
			fmt.Printf("Created user %s (ID: %d) with password 'Password123'\n", u.Username, uid)
		} else {
			_, _ = conn.Exec(ctx, "UPDATE users SET password = $1 WHERE id = $2", string(passwordHash), uid)
			fmt.Printf("User %s exists (ID: %d), password updated to 'Password123'\n", u.Username, uid)
		}

		// Check if conversation with danieljo17 exists
		var convID int64
		err = conn.QueryRow(ctx, `
			SELECT p1.conv_id 
			FROM participants p1
			JOIN participants p2 ON p1.conv_id = p2.conv_id
			WHERE p1.user_id = $1 AND p2.user_id = $2
			LIMIT 1
		`, danielID, uid).Scan(&convID)

		if err != nil {
			err = conn.QueryRow(ctx, "INSERT INTO conversations (created_at) VALUES (NOW()) RETURNING id").Scan(&convID)
			if err != nil {
				log.Printf("Failed to create conversation between %d and %d: %v", danielID, uid, err)
				continue
			}

			_, err = conn.Exec(ctx, "INSERT INTO participants (conv_id, user_id) VALUES ($1, $2), ($1, $3)", convID, danielID, uid)
			if err != nil {
				log.Printf("Failed to insert participants: %v", err)
				continue
			}
			fmt.Printf("Created conversation (ID: %d) between danieljo17 and %s\n", convID, u.Username)
		}

		// Insert messages if conversation has no messages
		var msgCount int64
		_ = conn.QueryRow(ctx, "SELECT COUNT(*) FROM messages WHERE conv_id = $1", convID).Scan(&msgCount)

		if msgCount == 0 {
			now := time.Now()
			for _, m := range u.Messages {
				senderID := danielID
				if m.FromUser {
					senderID = uid
				}
				msgTime := now.Add(-m.Offset)
				_, err = conn.Exec(
					ctx,
					"INSERT INTO messages (sender_id, conv_id, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $4)",
					senderID,
					convID,
					m.Content,
					msgTime,
				)
				if err != nil {
					log.Printf("Failed to insert message: %v", err)
				}
			}
			fmt.Printf("Populated %d messages for conversation %d\n", len(u.Messages), convID)
		}
	}

	fmt.Println("\n✓ Database seeding completed successfully!")
}
