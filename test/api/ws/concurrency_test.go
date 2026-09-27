package ws_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/stretchr/testify/mock"
)

// TestHub_ConcurrentOperations verifies that concurrent client registration, unregistration,
// direct message sending, read broadcasts, and slow client drops do not cause data races or panics.
func TestHub_ConcurrentOperations(t *testing.T) {
	mockService := new(MockConvService)
	mockService.On("GetContactUserIDs", mock.Anything, mock.Anything).Return([]int64{}, nil).Maybe()

	hub := ws.NewHub(mockService)
	go hub.Run()

	const numUsers = 20
	const numGoroutines = 10
	const duration = 1 * time.Second

	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	// Helper to create a client with a small buffer to intentionally trigger drops
	createClient := func(userID int64) *ws.Client {
		return &ws.Client{
			UserID: userID,
			Send:   make(chan []byte, 2),
			Hub:    hub,
		}
	}

	// 1. Goroutines registering and unregistering clients rapidly
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					userID := int64((id % numUsers) + 1)
					c := createClient(userID)
					hub.Register <- c
					time.Sleep(2 * time.Millisecond)
					hub.UnRegister <- c
				}
			}
		}(i)
	}

	// 2. Goroutines pumping SendDirect messages
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var msgID int64
			for {
				select {
				case <-stopCh:
					return
				default:
					msgID++
					senderID := int64((id % numUsers) + 1)
					recipientID := int64(((id + 1) % numUsers) + 1)
					hub.SendDirect <- &ws.DirectMessage{
						SenderID:       senderID,
						SenderUsername: fmt.Sprintf("user%d", senderID),
						RecipientID:    recipientID,
						ConvID:         100,
						Message: ws.WSMessage{
							ID:          msgID,
							SenderID:    senderID,
							RecipientID: recipientID,
							ConvID:      100,
							Content:     "concurrent message",
						},
						UnreadCount: 1,
					}
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(i)
	}

	// 3. Goroutines broadcasting read receipts
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-stopCh:
					return
				default:
					userID := int64((id % numUsers) + 1)
					hub.BroadcastRead <- &ws.ConversationReadPayload{
						ConvID:    100,
						UserID:    userID,
						MessageID: 1,
					}
					time.Sleep(2 * time.Millisecond)
				}
			}
		}(i)
	}

	// 4. Goroutines checking online status
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			userIDs := []int64{1, 2, 3, 4, 5}
			for {
				select {
				case <-stopCh:
					return
				default:
					_, _ = hub.GetOnlineStatus(context.Background(), userIDs)
					time.Sleep(5 * time.Millisecond)
				}
			}
		}()
	}

	time.Sleep(duration)
	close(stopCh)
	wg.Wait()
}

// TestClient_TrySendAndCloseConcurrency ensures concurrent TrySend and Close calls never panic.
func TestClient_TrySendAndCloseConcurrency(t *testing.T) {
	const iterations = 500
	for i := 0; i < iterations; i++ {
		client := &ws.Client{
			UserID: 1,
			Send:   make(chan []byte, 5),
		}

		var wg sync.WaitGroup
		wg.Add(3)

		// Goroutine 1: Continuous TrySend
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = client.TrySend([]byte("ping"))
			}
		}()

		// Goroutine 2: Continuous TrySend
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = client.TrySend([]byte("pong"))
			}
		}()

		// Goroutine 3: Concurrent Close
		go func() {
			defer wg.Done()
			time.Sleep(time.Duration(i%5) * time.Microsecond)
			client.Close()
		}()

		wg.Wait()
	}
}
