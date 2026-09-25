package ws_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockConvService struct {
	mock.Mock
}

func (m *MockConvService) GetContactUserIDs(ctx context.Context, userID int64) ([]int64, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func TestHub_GetOnlineStatus(t *testing.T) {
	mockService := new(MockConvService)
	hub := ws.NewHub(mockService)
	go hub.Run()

	// 1. Initially both users are offline
	statusMap, err := hub.GetOnlineStatus(context.Background(), []int64{102, 103})
	require.NoError(t, err)
	assert.False(t, statusMap[102])
	assert.False(t, statusMap[103])

	// 2. Register user 102
	mockService.On("GetContactUserIDs", mock.Anything, int64(102)).Return([]int64{}, nil).Maybe()

	client102 := &ws.Client{
		UserID: 102,
		Send:   make(chan []byte, 10),
		Hub:    hub,
	}
	hub.Register <- client102

	// Allow goroutine to process registration
	require.Eventually(t, func() bool {
		m, err := hub.GetOnlineStatus(context.Background(), []int64{102})
		return err == nil && m[102]
	}, 1*time.Second, 10*time.Millisecond)

	// 3. User 103 is still offline
	statusMap2, err := hub.GetOnlineStatus(context.Background(), []int64{102, 103})
	require.NoError(t, err)
	assert.True(t, statusMap2[102])
	assert.False(t, statusMap2[103])

	// 4. Unregister user 102
	hub.UnRegister <- client102
	require.Eventually(t, func() bool {
		m, err := hub.GetOnlineStatus(context.Background(), []int64{102})
		return err == nil && !m[102]
	}, 1*time.Second, 10*time.Millisecond)
}

func TestHub_PresenceSnapshotAndNotification(t *testing.T) {
	mockService := new(MockConvService)
	mockService.On("GetContactUserIDs", mock.Anything, int64(102)).Return([]int64{103}, nil).Maybe()
	mockService.On("GetContactUserIDs", mock.Anything, int64(103)).Return([]int64{102}, nil).Maybe()

	hub := ws.NewHub(mockService)
	go hub.Run()

	client102 := &ws.Client{
		UserID: 102,
		Send:   make(chan []byte, 10),
		Hub:    hub,
	}
	hub.Register <- client102

	// User 102 receives presence snapshot (peers are currently offline)
	select {
	case msg := <-client102.Send:
		var notif struct {
			Type    ws.WSNotificationType      `json:"type"`
			Payload ws.PresenceSnapshotPayload `json:"payload"`
		}
		err := json.Unmarshal(msg, &notif)
		require.NoError(t, err)
		assert.Equal(t, ws.TypePresenceSnapshot, notif.Type)
		assert.Empty(t, notif.Payload.OnlineUserIDs)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for presence snapshot")
	}

	// User 103 connects
	client103 := &ws.Client{
		UserID: 103,
		Send:   make(chan []byte, 10),
		Hub:    hub,
	}
	hub.Register <- client103

	// User 103 receives presence snapshot containing online peer 102
	select {
	case msg := <-client103.Send:
		var notif struct {
			Type    ws.WSNotificationType      `json:"type"`
			Payload ws.PresenceSnapshotPayload `json:"payload"`
		}
		err := json.Unmarshal(msg, &notif)
		require.NoError(t, err)
		assert.Equal(t, ws.TypePresenceSnapshot, notif.Type)
		assert.Contains(t, notif.Payload.OnlineUserIDs, int64(102))
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for 103 presence snapshot")
	}

	// User 102 receives user_presence (103 online)
	select {
	case msg := <-client102.Send:
		var notif struct {
			Type    ws.WSNotificationType  `json:"type"`
			Payload ws.UserPresencePayload `json:"payload"`
		}
		err := json.Unmarshal(msg, &notif)
		require.NoError(t, err)
		assert.Equal(t, ws.TypeUserPresence, notif.Type)
		assert.Equal(t, int64(103), notif.Payload.UserID)
		assert.True(t, notif.Payload.Online)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for presence update")
	}
}

func TestHub_PresenceLifecycle_DanielAlwaysOnline(t *testing.T) {
	danielID := int64(1)
	ethanID := int64(2)

	mockService := new(MockConvService)
	mockService.On("GetContactUserIDs", mock.Anything, danielID).Return([]int64{ethanID}, nil).Maybe()
	mockService.On("GetContactUserIDs", mock.Anything, ethanID).Return([]int64{danielID}, nil).Maybe()

	hub := ws.NewHub(mockService)
	go hub.Run()

	// Daniel connects and STAYS online throughout the test
	danielClient := &ws.Client{UserID: danielID, Send: make(chan []byte, 20), Hub: hub}
	hub.Register <- danielClient

	select {
	case <-danielClient.Send:
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Daniel's initial snapshot")
	}

	// === Cycle 1: Ethan connects, sees Daniel online, then disconnects ===
	ethanClient1 := &ws.Client{UserID: ethanID, Send: make(chan []byte, 10), Hub: hub}
	hub.Register <- ethanClient1

	// Ethan's snapshot must contain Daniel
	select {
	case msg := <-ethanClient1.Send:
		var notif struct {
			Type    ws.WSNotificationType      `json:"type"`
			Payload ws.PresenceSnapshotPayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypePresenceSnapshot, notif.Type)
		assert.Equal(t, []int64{danielID}, notif.Payload.OnlineUserIDs)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Ethan's cycle 1 snapshot")
	}

	// Daniel receives real-time user_presence (Ethan online: true)
	select {
	case msg := <-danielClient.Send:
		var notif struct {
			Type    ws.WSNotificationType  `json:"type"`
			Payload ws.UserPresencePayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypeUserPresence, notif.Type)
		assert.Equal(t, ethanID, notif.Payload.UserID)
		assert.True(t, notif.Payload.Online)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Daniel's cycle 1 presence update")
	}

	// Ethan disconnects
	hub.UnRegister <- ethanClient1

	// Daniel receives real-time user_presence (Ethan online: false)
	select {
	case msg := <-danielClient.Send:
		var notif struct {
			Type    ws.WSNotificationType  `json:"type"`
			Payload ws.UserPresencePayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypeUserPresence, notif.Type)
		assert.Equal(t, ethanID, notif.Payload.UserID)
		assert.False(t, notif.Payload.Online)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Daniel's cycle 1 offline update")
	}

	time.Sleep(1 * time.Second)

	// === Cycle 2: Ethan reconnects, still sees Daniel online, then disconnects ===
	ethanClient2 := &ws.Client{UserID: ethanID, Send: make(chan []byte, 10), Hub: hub}
	hub.Register <- ethanClient2

	// Ethan's snapshot must STILL contain Daniel
	select {
	case msg := <-ethanClient2.Send:
		var notif struct {
			Type    ws.WSNotificationType      `json:"type"`
			Payload ws.PresenceSnapshotPayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypePresenceSnapshot, notif.Type)
		assert.Equal(t, []int64{danielID}, notif.Payload.OnlineUserIDs)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Ethan's cycle 2 snapshot")
	}

	// Daniel receives real-time user_presence (Ethan online: true)
	select {
	case msg := <-danielClient.Send:
		var notif struct {
			Type    ws.WSNotificationType  `json:"type"`
			Payload ws.UserPresencePayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypeUserPresence, notif.Type)
		assert.Equal(t, ethanID, notif.Payload.UserID)
		assert.True(t, notif.Payload.Online)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Daniel's cycle 2 presence update")
	}

	// Ethan disconnects again
	hub.UnRegister <- ethanClient2

	// Daniel receives real-time user_presence (Ethan online: false)
	select {
	case msg := <-danielClient.Send:
		var notif struct {
			Type    ws.WSNotificationType  `json:"type"`
			Payload ws.UserPresencePayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypeUserPresence, notif.Type)
		assert.Equal(t, ethanID, notif.Payload.UserID)
		assert.False(t, notif.Payload.Online)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Daniel's cycle 2 offline update")
	}

	time.Sleep(1 * time.Second)

	// === Cycle 3: Ethan connects a 3rd time, still sees Daniel online ===
	ethanClient3 := &ws.Client{UserID: ethanID, Send: make(chan []byte, 10), Hub: hub}
	hub.Register <- ethanClient3

	// Ethan's snapshot must STILL contain Daniel
	select {
	case msg := <-ethanClient3.Send:
		var notif struct {
			Type    ws.WSNotificationType      `json:"type"`
			Payload ws.PresenceSnapshotPayload `json:"payload"`
		}
		require.NoError(t, json.Unmarshal(msg, &notif))
		assert.Equal(t, ws.TypePresenceSnapshot, notif.Type)
		assert.Equal(t, []int64{danielID}, notif.Payload.OnlineUserIDs)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for Ethan's cycle 3 snapshot")
	}
}
