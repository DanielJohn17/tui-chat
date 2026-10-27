package chat

type FetchStatus int

const (
	StatusIdle FetchStatus = iota
	StatusLoading
	StatusSuccess
	StatusFailed
)

type ChatTab int

const (
	TabCloud ChatTab = iota
	TabSecret
)

type ToggleTabMsg struct{}
