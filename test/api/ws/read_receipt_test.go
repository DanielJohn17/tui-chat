package ws_test

import (
	"testing"
	"time"

	"github.com/DanielJohn17/tui-chat/internal/api/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ReadReceiptWorker_Coalescing(t *testing.T) {
	persister := &mockPersister{}
	hub := ws.NewHub(nil)
	client := ws.NewClient(42, "tester", nil, hub, persister)

	client.StartReadReceiptWorker()
	defer client.Close()

	// Drain hub broadcast channel in background so it doesn't back up
	go func() {
		for range hub.BroadcastRead {
		}
	}()

	// Rapidly enqueue 100 read receipts for conversation 1 with sequential IDs
	for i := int64(1); i <= 100; i++ {
		client.EnqueueReadReceipt(ws.MarkReadPayload{
			ConvID:    1,
			MessageID: i,
		})
	}

	// Verify that the worker reaches the highest message ID (100)
	require.Eventually(t, func() bool {
		calls := persister.getCalls()
		if len(calls) == 0 {
			return false
		}
		// The latest processed call for conv 1 must be 100
		lastCall := calls[len(calls)-1]
		return lastCall.ConvID == 1 && lastCall.MessageID == 100
	}, 2*time.Second, 10*time.Millisecond)

	calls := persister.getCalls()
	// Because of coalescing, total DB calls should be substantially less than 100
	assert.Less(t, len(calls), 100, "expected rapid receipts to be coalesced into fewer DB calls")

	// Sending an older or equal message ID should be filtered out by monotonic watermark check
	priorLen := len(calls)
	client.EnqueueReadReceipt(ws.MarkReadPayload{
		ConvID:    1,
		MessageID: 50,
	})
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, priorLen, len(persister.getCalls()), "older receipt should have been filtered without DB call")
}

func TestClient_EnqueueReadReceipt_NonBlockingOnClosedOrNil(t *testing.T) {
	// A client with nil channels should not panic or block
	clientNil := &ws.Client{UserID: 1}
	clientNil.EnqueueReadReceipt(ws.MarkReadPayload{ConvID: 1, MessageID: 10})

	// A closed client should discard without blocking
	hub := ws.NewHub(nil)
	persister := &mockPersister{}
	client := ws.NewClient(10, "alice", nil, hub, persister)
	client.Close()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 200; i++ {
			client.EnqueueReadReceipt(ws.MarkReadPayload{ConvID: 1, MessageID: int64(i)})
		}
		close(done)
	}()

	select {
	case <-done:
		// success: did not hang
	case <-time.After(500 * time.Millisecond):
		t.Fatal("EnqueueReadReceipt blocked on closed client")
	}
}

func TestClient_ReadReceiptWorker_MultiConversationBatch(t *testing.T) {
	persister := &mockPersister{}
	hub := ws.NewHub(nil)
	client := ws.NewClient(77, "bob", nil, hub, persister)

	client.StartReadReceiptWorker()
	defer client.Close()

	go func() {
		for range hub.BroadcastRead {
		}
	}()

	// Interleave receipts across two distinct conversations
	for i := int64(1); i <= 50; i++ {
		client.EnqueueReadReceipt(ws.MarkReadPayload{ConvID: 10, MessageID: i})
		client.EnqueueReadReceipt(ws.MarkReadPayload{ConvID: 20, MessageID: i + 100})
	}

	require.Eventually(t, func() bool {
		calls := persister.getCalls()
		var foundConv10, foundConv20 bool
		for _, c := range calls {
			if c.ConvID == 10 && c.MessageID == 50 {
				foundConv10 = true
			}
			if c.ConvID == 20 && c.MessageID == 150 {
				foundConv20 = true
			}
		}
		return foundConv10 && foundConv20
	}, 2*time.Second, 10*time.Millisecond)
}
