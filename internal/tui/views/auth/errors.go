package auth

import "strings"

// CleanAuthError translates backend validator and server errors into human-friendly, actionable messages.
func CleanAuthError(err error) string {
	if err == nil {
		return ""
	}
	return CleanAuthErrorString(err.Error())
}

// CleanAuthErrorString cleans raw error message strings.
func CleanAuthErrorString(msg string) string {
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)

	// Specific backend auth service errors
	if strings.Contains(lower, "incorrect username or password") {
		return "Incorrect username or password. Please try again."
	}
	if strings.Contains(lower, "user already exists") {
		return "Username is already taken. Please choose another username."
	}

	// Go validator/v10 tag errors
	if strings.Contains(trimmed, "containsany") {
		return "Password must contain at least one uppercase letter (A-Z) and one number (0-9)."
	}
	if strings.Contains(trimmed, "validation error") {
		if strings.Contains(trimmed, "Password") {
			return "Password must be at least 6 characters and contain uppercase letters and numbers."
		}
		if strings.Contains(trimmed, "Username") {
			return "Username must be between 3 and 20 characters."
		}
		if strings.Contains(trimmed, "Name") {
			return "Display name must be between 3 and 50 characters."
		}
		return "Please verify that all fields are filled out correctly."
	}
	if strings.Contains(trimmed, "failed on the 'min' tag") {
		if strings.Contains(trimmed, "Username") {
			return "Username must be at least 3 characters."
		}
		if strings.Contains(trimmed, "Name") {
			return "Display name must be at least 3 characters."
		}
		if strings.Contains(trimmed, "Password") {
			return "Password must be at least 6 characters."
		}
		return "One or more fields do not meet the minimum length requirement."
	}
	if strings.Contains(trimmed, "failed on the 'max' tag") {
		return "One or more fields exceed the maximum length."
	}
	if strings.Contains(trimmed, "failed on the 'required' tag") {
		return "Please fill in all required fields."
	}

	// Server / network issues
	if strings.Contains(lower, "offline") || strings.Contains(lower, "unreachable") || strings.Contains(lower, "connection refused") {
		return "Chat server is currently offline or unreachable."
	}
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "timed out") {
		return "Request timed out. Please try again."
	}
	if strings.Contains(lower, "http 500") || strings.Contains(lower, "internal server error") {
		return "Server error occurred. Please try again later."
	}
	if strings.Contains(lower, "http 401") || strings.Contains(lower, "unauthorized") {
		return "Invalid credentials. Please verify your username and password."
	}

	// Check if this looks like an internal Go or SQL error
	if strings.Contains(trimmed, "Key:") || strings.Contains(trimmed, "Error:") || strings.Contains(lower, "sql") || strings.Contains(lower, "syntax") {
		return "Invalid input provided. Please verify your information and try again."
	}

	return trimmed
}
