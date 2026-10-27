package auth

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) updateFocus() {
	if m.mode == ModeLogin {
		if m.focusIndex == 0 {
			m.loginUsername.Focus()
			m.loginPassword.Blur()
		} else {
			m.loginUsername.Blur()
			m.loginPassword.Focus()
		}
	} else {
		m.regName.Blur()
		m.regUsername.Blur()
		m.regPassword.Blur()
		switch m.focusIndex {
		case 0:
			m.regName.Focus()
		case 1:
			m.regUsername.Focus()
		case 2:
			m.regPassword.Focus()
		}
	}
}

func (m *Model) submit() tea.Cmd {
	if m.mode == ModeLogin {
		return m.submitLogin()
	}
	return m.submitRegister()
}

func (m *Model) submitLogin() tea.Cmd {
	username := strings.TrimSpace(m.loginUsername.Value())
	password := strings.TrimSpace(m.loginPassword.Value())

	if username == "" {
		m.errorMessage = "Please enter your username"
		return nil
	}
	if len(username) < 3 {
		m.errorMessage = "Username must be at least 3 characters"
		return nil
	}
	if len(username) > 20 {
		m.errorMessage = "Username must not exceed 20 characters"
		return nil
	}
	if password == "" {
		m.errorMessage = "Please enter your password"
		return nil
	}

	m.loading = true
	m.errorMessage = ""
	c := m.client

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			prof, err := c.Login(username, password)
			if err != nil {
				return AuthErrorMsg{Err: err}
			}
			return AuthSuccessMsg{Profile: *prof}
		},
	)
}

func (m *Model) submitRegister() tea.Cmd {
	name := strings.TrimSpace(m.regName.Value())
	username := strings.TrimSpace(m.regUsername.Value())
	password := strings.TrimSpace(m.regPassword.Value())

	if name == "" {
		m.errorMessage = "Please enter your display name"
		return nil
	}
	if len(name) < 3 {
		m.errorMessage = "Display name must be at least 3 characters"
		return nil
	}
	if len(name) > 50 {
		m.errorMessage = "Display name must not exceed 50 characters"
		return nil
	}
	if username == "" {
		m.errorMessage = "Please enter a username"
		return nil
	}
	if len(username) < 3 {
		m.errorMessage = "Username must be at least 3 characters"
		return nil
	}
	if len(username) > 20 {
		m.errorMessage = "Username must not exceed 20 characters"
		return nil
	}
	if password == "" {
		m.errorMessage = "Please enter a password"
		return nil
	}
	if len(password) < 6 {
		m.errorMessage = "Password must be at least 6 characters"
		return nil
	}
	if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		m.errorMessage = "Password must contain at least one capital letter (A-Z)"
		return nil
	}
	if !strings.ContainsAny(password, "0123456789") {
		m.errorMessage = "Password must contain at least one number (0-9)"
		return nil
	}

	m.loading = true
	m.errorMessage = ""
	c := m.client

	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			prof, err := c.Register(name, username, password)
			if err != nil {
				return AuthErrorMsg{Err: err}
			}
			return AuthSuccessMsg{Profile: *prof}
		},
	)
}
