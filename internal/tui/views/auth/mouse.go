package auth

import tea "github.com/charmbracelet/bubbletea"

func (m *Model) handleMouse(msg tea.MouseMsg) (Model, tea.Cmd) {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return *m, nil
	}

	cardWidth := 56
	if m.width > 0 && m.width < 60 {
		cardWidth = m.width - 4
	}
	if cardWidth < 44 {
		cardWidth = 44
	}
	cardW := cardWidth + 6

	cardH := 25
	if m.mode == ModeRegister {
		cardH = 33
	}
	if m.errorMessage != "" {
		cardH += 4
	}

	startX := (m.width - cardW) / 2
	startY := (m.height - cardH) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	relX := msg.X - startX
	relY := msg.Y - startY

	if relX < 0 || relX >= cardW {
		return *m, nil
	}

	// 1. Tabs row (lines 6-7)
	if relY >= 6 && relY <= 7 {
		if relX < cardW/2 {
			m.mode = ModeLogin
		} else {
			m.mode = ModeRegister
		}
		m.focusIndex = 0
		m.updateFocus()
		return *m, nil
	}

	errorOffset := 0
	if m.errorMessage != "" {
		errorOffset = 4
	}

	if m.mode == ModeLogin {
		// Username field (lines 9-12 + errorOffset)
		if relY >= 9+errorOffset && relY <= 12+errorOffset {
			m.focusIndex = 0
			m.updateFocus()
			return *m, nil
		}
		// Password field (lines 14-17 + errorOffset)
		if relY >= 14+errorOffset && relY <= 17+errorOffset {
			m.focusIndex = 1
			m.updateFocus()
			return *m, nil
		}
		// Submit button (lines 18-20 + errorOffset)
		if relY >= 18+errorOffset && relY <= 20+errorOffset {
			if !m.loading {
				return *m, m.submit()
			}
			return *m, nil
		}
	} else {
		// Register mode
		// Display Name field (lines 9-12 + errorOffset)
		if relY >= 9+errorOffset && relY <= 12+errorOffset {
			m.focusIndex = 0
			m.updateFocus()
			return *m, nil
		}
		// Username field (lines 14-17 + errorOffset)
		if relY >= 14+errorOffset && relY <= 17+errorOffset {
			m.focusIndex = 1
			m.updateFocus()
			return *m, nil
		}
		// Password field (lines 19-22 + errorOffset)
		if relY >= 19+errorOffset && relY <= 22+errorOffset {
			m.focusIndex = 2
			m.updateFocus()
			return *m, nil
		}
		// Submit button (lines 25-28 + errorOffset)
		if relY >= 25+errorOffset && relY <= 28+errorOffset {
			if !m.loading {
				return *m, m.submit()
			}
			return *m, nil
		}
	}

	return *m, nil
}
