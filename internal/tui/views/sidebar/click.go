package sidebar

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
)

type ClickableItem struct {
	ChatIndex int
	StartY    int
	EndY      int
}

// HandleClick evaluates click coordinates relative to the sidebar:
// clickedChat: pointer to selected chat if a chat row was clicked, else nil
// clickedNewDM: true if [+n] was clicked
// clickedProfile: true if bottom profile card was clicked
func (m *Model) HandleClick(relX, relY int) (*client.Chat, bool, bool) {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return nil, false, false
	}

	// If clickableItems not yet initialized, render once to populate positions
	if len(m.clickableItems) == 0 {
		_ = m.View()
	}

	// Line 0: top border; Line 1: Header [+n]
	if relY <= 1 {
		if relX >= m.width-10 {
			return nil, true, false
		}
		return nil, false, false
	}

	// Top scroll indicator clicked
	if m.topScrollY != -1 && relY == m.topScrollY {
		m.MoveUp()
		return m.SelectedChat(), false, false
	}

	// Bottom scroll indicator clicked
	if m.bottomScrollY != -1 && relY == m.bottomScrollY {
		m.MoveDown()
		return m.SelectedChat(), false, false
	}

	// Bottom 4 lines: divider + profile info + profile hint + bottom border
	if relY >= m.height-4 && relY < m.height {
		return nil, false, true
	}

	// Match against accurately recorded clickable items from render
	for _, item := range m.clickableItems {
		if relY >= item.StartY && relY <= item.EndY {
			if item.ChatIndex >= 0 && item.ChatIndex < len(chats) {
				m.selectedIndex = item.ChatIndex
				m.ensureVisible()
				return &chats[item.ChatIndex], false, false
			}
		}
	}

	return nil, false, false
}
