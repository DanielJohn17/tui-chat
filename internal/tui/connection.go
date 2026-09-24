package tui

import (
	"strings"
	"time"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/statusbar"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleWSConnectResult(msg WSConnectResultMsg) (tea.Model, tea.Cmd) {
	if msg.Err != nil {
		errStr := strings.ToLower(msg.Err.Error())
		if strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "invalid token") {
			_ = m.client.Logout()
			m.state = StateAuth
			m.authView.SetErrorMessage("Session expired. Please log in again.")
			return *m, nil
		}
		m.connStatus = statusbar.StatusReconnecting
		delay := m.reconnectDelay
		if delay == 0 {
			delay = time.Second
		}
		m.reconnectDelay = delay * 2
		if m.reconnectDelay > 15*time.Second {
			m.reconnectDelay = 15 * time.Second
		}
		m.retrySecondsLeft = int(delay.Seconds())
		if m.retrySecondsLeft <= 0 {
			m.retrySecondsLeft = 1
		}
		return *m, tea.Batch(
			tea.Tick(time.Second, func(t time.Time) tea.Msg { return RetryCountdownTickMsg{} }),
			tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
		)
	}

	m.connStatus = statusbar.StatusConnected
	m.reconnectDelay = time.Second
	m.retrySecondsLeft = 0

	var postConnectCmds []tea.Cmd
	postConnectCmds = append(postConnectCmds, fetchChatsCmd(m.client))
	if activeID := m.chatView.ActiveChatID(); activeID != 0 {
		_ = m.client.FocusConv(activeID)
		postConnectCmds = append(postConnectCmds, fetchMessagesCmd(m.client, activeID))
	}
	return *m, tea.Batch(postConnectCmds...)
}

func (m *AppModel) handleRetryCountdownTick() (tea.Model, tea.Cmd) {
	if m.connStatus == statusbar.StatusReconnecting {
		m.retrySecondsLeft--
		if m.retrySecondsLeft > 0 {
			return *m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return RetryCountdownTickMsg{} })
		}
		m.connStatus = statusbar.StatusConnecting
		return *m, tea.Batch(
			connectWSCmd(m.client, m.wsEventsChan),
			tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
		)
	}
	return *m, nil
}

func (m *AppModel) handleSpinnerTick() (tea.Model, tea.Cmd) {
	if m.connStatus == statusbar.StatusConnecting || m.connStatus == statusbar.StatusReconnecting {
		m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
		return *m, tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} })
	}
	return *m, nil
}

func (m *AppModel) handleReconnectTick() (tea.Model, tea.Cmd) {
	if m.state == StateChat && m.client.IsAuthenticated() {
		if !m.client.IsWSConnected() {
			m.connStatus = statusbar.StatusConnecting
			return *m, connectWSCmd(m.client, m.wsEventsChan)
		}
		m.connStatus = statusbar.StatusConnected
	}
	return *m, nil
}

func (m *AppModel) triggerManualReconnect() (tea.Model, tea.Cmd) {
	if m.connStatus != statusbar.StatusConnected {
		m.connStatus = statusbar.StatusConnecting
		m.reconnectDelay = time.Second
		m.retrySecondsLeft = 0
		return *m, tea.Batch(
			connectWSCmd(m.client, m.wsEventsChan),
			tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
		)
	}
	return *m, nil
}

func (m *AppModel) scheduleDisconnectRetry() tea.Cmd {
	m.connStatus = statusbar.StatusReconnecting
	delay := m.reconnectDelay
	if delay == 0 {
		delay = time.Second
	}
	m.reconnectDelay = delay * 2
	if m.reconnectDelay > 15*time.Second {
		m.reconnectDelay = 15 * time.Second
	}
	m.retrySecondsLeft = int(delay.Seconds())
	if m.retrySecondsLeft <= 0 {
		m.retrySecondsLeft = 1
	}
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg { return RetryCountdownTickMsg{} }),
		tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return SpinnerTickMsg{} }),
	)
}
