package sidebar

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	client        client.Client
	selectedIndex int
	scrollOffset  int
	width         int
	height        int

	// Accurate click targets recorded on each render
	clickableItems []ClickableItem
	topScrollY     int
	bottomScrollY  int
}

func New(c client.Client) Model {
	return Model{
		client:        c,
		selectedIndex: 0,
		scrollOffset:  0,
		width:         32,
		height:        24,
		topScrollY:    -1,
		bottomScrollY: -1,
	}
}

func (m *Model) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Model) ensureVisible() {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
	if m.selectedIndex >= len(chats) {
		m.selectedIndex = len(chats) - 1
	}

	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	}

	availableListHeight := m.height - 7
	if availableListHeight < 4 {
		availableListHeight = 4
	}

	// Effective max lines available for items taking scroll indicators into account
	maxItemsHeight := availableListHeight
	if m.scrollOffset > 0 {
		maxItemsHeight--
	}
	if m.selectedIndex < len(chats)-1 {
		maxItemsHeight--
	}
	if maxItemsHeight < 4 {
		maxItemsHeight = 4
	}

	// Each item takes 3 lines + 1 separator line = 4 lines (first item doesn't have leading separator)
	count := m.selectedIndex - m.scrollOffset + 1
	lines := count*4 - 1
	if lines < 0 {
		lines = 0
	}

	for lines > maxItemsHeight && m.scrollOffset < m.selectedIndex {
		lines -= 4
		m.scrollOffset++
	}
}

func (m *Model) MoveUp() {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return
	}
	if m.selectedIndex > 0 {
		m.selectedIndex--
		m.ensureVisible()
	}
}

func (m *Model) MoveDown() {
	chats := m.client.Chats()
	if len(chats) == 0 {
		return
	}
	if m.selectedIndex < len(chats)-1 {
		m.selectedIndex++
		m.ensureVisible()
	}
}

func (m *Model) SelectIndex(idx int) {
	chats := m.client.Chats()
	if idx >= 0 && idx < len(chats) {
		m.selectedIndex = idx
		m.ensureVisible()
	}
}

func (m Model) SelectedIndex() int {
	return m.selectedIndex
}

func (m Model) SelectedChat() *client.Chat {
	chats := m.client.Chats()
	if len(chats) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(chats) {
		return nil
	}
	return &chats[m.selectedIndex]
}

func (m *Model) View() string {
	chats := m.client.Chats()
	contentWidth := m.width - 4
	if contentWidth < 20 {
		contentWidth = 20
	}

	// 1. Header component
	header := renderHeader(contentWidth)

	// Available inner height inside rounded border
	innerH := m.height - 2
	if innerH < 8 {
		innerH = 8
	}

	headerHeight := 2
	footerHeight := 3
	availableListHeight := innerH - headerHeight - footerHeight
	if availableListHeight < 4 {
		availableListHeight = 4
	}

	m.ensureVisible()

	// 2. Chat list items with bounded window rendering
	var chatItems []string
	m.clickableItems = nil
	m.topScrollY = -1
	m.bottomScrollY = -1

	currentY := 3
	hasTopScroll := m.scrollOffset > 0
	if hasTopScroll {
		m.topScrollY = currentY
		currentY++
	}

	usedLines := 0
	if hasTopScroll {
		usedLines++
	}

	for i := m.scrollOffset; i < len(chats); i++ {
		ch := chats[i]
		isSelected := i == m.selectedIndex

		itemBox := renderChatItem(ch, isSelected, contentWidth)
		actualLines := lipgloss.Height(itemBox)

		linesNeeded := actualLines
		if len(chatItems) > 0 {
			linesNeeded += 1 // 1 separator line before this item
		}
		if i < len(chats)-1 {
			linesNeeded++ // reserve 1 line for bottom scroll indicator
		}

		if usedLines+linesNeeded > availableListHeight && len(chatItems) > 0 {
			break
		}

		if len(chatItems) > 0 {
			if len(m.clickableItems) > 0 {
				m.clickableItems[len(m.clickableItems)-1].EndY++
			}
			chatItems = append(chatItems, "")
			currentY++
			usedLines++
		}

		m.clickableItems = append(m.clickableItems, ClickableItem{
			ChatIndex: i,
			StartY:    currentY,
			EndY:      currentY + actualLines - 1,
		})
		currentY += actualLines
		usedLines += actualLines
		chatItems = append(chatItems, itemBox)
	}

	// Add scroll indicators if items exist above or below
	if hasTopScroll && len(chatItems) > 0 {
		chatItems[0] = theme.StyleDim.Render("   ▲ more above") + "\n" + chatItems[0]
	}
	if m.scrollOffset+len(m.clickableItems) < len(chats) && len(chatItems) > 0 {
		m.bottomScrollY = currentY
		lastIdx := len(chatItems) - 1
		chatItems[lastIdx] = chatItems[lastIdx] + "\n" + theme.StyleDim.Render("   ▼ more below")
	}

	chatsList := lipgloss.JoinVertical(lipgloss.Left, chatItems...)

	// 3. User Profile Footer component pinned at the bottom
	footerBlock := renderFooter(m.client.Profile(), contentWidth)

	// Calculate blank filler lines between chat list and footer to pin footer to bottom
	renderedChatsHeight := lipgloss.Height(chatsList)
	fillerLinesCount := availableListHeight - renderedChatsHeight
	if fillerLinesCount < 0 {
		fillerLinesCount = 0
	}

	// Assemble sidebar container
	bodyItems := []string{
		header,
		"",
		chatsList,
	}
	for k := 0; k < fillerLinesCount; k++ {
		bodyItems = append(bodyItems, "")
	}
	bodyItems = append(bodyItems, footerBlock)

	boxWidth := m.width - 2
	if boxWidth < 10 {
		boxWidth = 10
	}
	boxHeight := m.height - 2
	if boxHeight < 6 {
		boxHeight = 6
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.ColorBorderDim).
		Width(boxWidth).
		Height(boxHeight).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, bodyItems...))
}
