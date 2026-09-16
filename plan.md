# Implementation Plan - Telegram-Style Sidebar & Profile Mouse Click Fixes

Refactor the sidebar chat list to match Telegram's two-line layout without expansion on focus, implement unseen message notification badges and mock interactions, and fix mouse click hit-testing in the profile edit page.

## User Review Required

> [!NOTE]
> As requested by the user, the test color-coding item has been removed from this plan.

> [!IMPORTANT]
> **Sidebar Layout Shift**: Previously, clicking an item expanded its height from 2 lines to 3 lines (revealing `LastMessage`). In the new Telegram-style layout:
> 1. Both selected and unselected items will always occupy **2 lines** (fixed height, no expand on focus).
> 2. Line 1: Status indicator + Name (left), Timestamp (right).
> 3. Line 2: Last message sent/received preview (left), Unseen message notification badge pill (right).
> 4. Active/selected item is distinguished with an active highlight style (cyan accent / marker / row background) without altering its height.

---

## Proposed Changes

### Component 1: Sidebar Views (`app/internal/tui/views/sidebar/`)

#### [MODIFY] item.go
- Redesign `renderChatItem(ch client.Chat, isSelected bool, contentWidth int)`:
  - **Line 1 (Top)**:
    - Left: Indicator (`▶ ` with bold title for selected, `● ` or `○ ` with name for unselected).
    - Right: Timestamp (`ch.Time`).
  - **Line 2 (Bottom)**:
    - Left: Preview of last message sent/received (`theme.Truncate(ch.LastMessage, ...)`), styled in dim or highlighted text.
    - Right: Unseen message badge pill if `ch.Unread > 0` (Telegram-style padded badge pill, e.g. ` 2 `, ` 188 `).
  - Selected state uses active cyan styling and consistent 2-line height (no vertical expansion).
  - Proper padding within `contentWidth`.

#### [MODIFY] sidebar.go
- Simplify `ensureVisible()` to rely on uniform 2-line item heights (`itemHeight = 2`), guaranteeing 100% predictable scrolling.
- Ensure hit-test ranges in `clickableItems` record exact `StartY` and `EndY` for each 2-line item.

---

### Component 2: Client & Mock Notifications (`app/internal/tui/client/`)

#### [MODIFY] client.go
- Add `MarkRead(chatID int64)` to `Client` interface so selecting/opening a chat can clear unseen message notifications.

#### [MODIFY] mock.go
- Implement `MarkRead(chatID int64)` on `mockClient` to reset `Unread = 0`.
- Populate mock chats with rich unseen message counts (e.g. 1, 2, 3, 5, 12) to showcase the Telegram-style notification pills.

---

### Component 3: Profile Page Mouse Clicks (`app/internal/tui/views/profile/`)

#### [MODIFY] profile.go
- Fix mouse hit-testing in `Update`:
  - Calculate actual `cardW` and `cardH` from rendered card dimensions rather than hardcoded 24 and 60.
  - Dynamically compute `startX` and `startY` matching `lipgloss.Place(...)`.
  - Fix row hit-testing for:
    - Display Name input (focus index 0)
    - Username input (focus index 1)
    - Bio / Status input (focus index 2)
    - Save Changes button (action index 3 / save)
    - Back to Chat button (action index 4 / back)
  - Account for alert box presence so field hit-testing never drifts when validation errors appear.

---

### Component 4: App Integration & Tests (`app/internal/tui/` and `app/test/`)

#### [MODIFY] app.go
- When a chat is clicked in `sidebarView`, call `m.client.MarkRead(clickedChat.ID)` to clear its unseen message badge.

#### [MODIFY] ui_enhancements_test.go
- Update existing tests to reflect the new Telegram 2-line format (no expand on focus).
- Add new unit tests:
  - `TestTelegramSidebarLayout`: Verifies 2-line fixed height, last message preview on all items, and unseen message pill rendering.
  - `TestUnseenMessageMarkRead`: Verifies clicking a chat with unread messages marks them read and clears the badge.
  - `TestProfilePageMouseClicks`: Verifies clicking each input box (Name, Username, Bio) focuses that input, and clicking Save/Cancel executes the respective action.

---

## Verification Plan

### Automated Tests
- Run full test suite:
  ```bash
  go test -v ./...
  ```
- Run targeted TUI enhancement tests:
  ```bash
  go test -v -run "TestTelegramSidebarLayout|TestUnseenMessageMarkRead|TestProfilePageMouseClicks" ./test
  ```

### Manual Verification
- Validate visual formatting against the Telegram reference image:
  - Line 1: Name + timestamp
  - Line 2: Last message + unread badge
  - Selecting different chats keeps height constant without row jumping
  - Clicking on Display Name, Username, and Status in profile focuses the respective input field
  - Clicking Save/Cancel in profile triggers save/cancel without misfiring
