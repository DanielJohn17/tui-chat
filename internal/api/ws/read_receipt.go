package ws

import (
	"context"
	"log/slog"
	"time"
)

// EnqueueReadReceipt non-blockingly enqueues a read receipt to the client's dedicated worker.
// If the buffer is full, intermediate receipts are safely dropped without blocking the caller.
func (c *Client) EnqueueReadReceipt(payload MarkReadPayload) {
	if c.readReceipts == nil {
		return
	}
	if c.done != nil {
		select {
		case <-c.done:
			return
		default:
		}
	}

	if c.done != nil {
		select {
		case <-c.done:
			return
		case c.readReceipts <- payload:
		default:
			slog.Warn("client read receipts queue full, dropped intermediate receipt",
				"conv", payload.ConvID, "user", c.UserID)
		}
	} else {
		select {
		case c.readReceipts <- payload:
		default:
			slog.Warn("client read receipts queue full, dropped intermediate receipt",
				"conv", payload.ConvID, "user", c.UserID)
		}
	}
}

// StartReadReceiptWorker launches the dedicated read receipt coalescing worker goroutine.
func (c *Client) StartReadReceiptWorker() {
	go c.readReceiptWorker()
}

// readReceiptWorker sequentially drains, coalesces, and writes read watermarks.
func (c *Client) readReceiptWorker() {
	if c.readReceipts == nil {
		return
	}
	watermarks := make(map[int64]int64)

	for {
		var receipt MarkReadPayload
		var ok bool

		if c.done != nil {
			select {
			case <-c.done:
				return
			case receipt, ok = <-c.readReceipts:
				if !ok {
					return
				}
			}
		} else {
			receipt, ok = <-c.readReceipts
			if !ok {
				return
			}
		}

		// Coalesce burst of receipts in queue
		pending := make(map[int64]int64)
		pending[receipt.ConvID] = receipt.MessageID

	drain:
		for {
			if c.done != nil {
				select {
				case <-c.done:
					return
				case next, ok := <-c.readReceipts:
					if !ok {
						break drain
					}
					if currID, exists := pending[next.ConvID]; !exists || next.MessageID > currID || next.MessageID == 0 {
						pending[next.ConvID] = next.MessageID
					}
				default:
					break drain
				}
			} else {
				select {
				case next, ok := <-c.readReceipts:
					if !ok {
						break drain
					}
					if currID, exists := pending[next.ConvID]; !exists || next.MessageID > currID || next.MessageID == 0 {
						pending[next.ConvID] = next.MessageID
					}
				default:
					break drain
				}
			}
		}

		for convID, messageID := range pending {
			if c.done != nil {
				select {
				case <-c.done:
					return
				default:
				}
			}

			if lastID, exists := watermarks[convID]; exists && messageID > 0 && messageID <= lastID {
				continue
			}
			if messageID > 0 {
				watermarks[convID] = messageID
			}

			// 1. Update DB watermark
			if c.convService != nil {
				dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				c.convService.MarkAsRead(dbCtx, messageID, c.UserID, convID)
				cancel()
			}

			// 2. Hub broadcast
			if c.Hub != nil && c.Hub.BroadcastRead != nil {
				select {
				case c.Hub.BroadcastRead <- &ConversationReadPayload{
					ConvID:    convID,
					UserID:    c.UserID,
					MessageID: messageID,
				}:
				default:
					slog.Warn("hub BroadcastRead buffer full, dropped read receipt", "conv", convID)
				}
			}
		}
	}
}
