package models

import (
	"testing"
)

func TestBookStatusConstants(t *testing.T) {
	tests := []struct {
		status   BookStatus
		expected string
	}{
		{StatusWantToRead, "want-to-read"},
		{StatusReading, "reading"},
		{StatusFinished, "finished"},
	}

	for _, tt := range tests {
		if string(tt.status) != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status)
		}
	}
}

func TestBookStatusIsValid(t *testing.T) {
	validStatuses := []BookStatus{StatusWantToRead, StatusReading, StatusFinished}

	for _, status := range validStatuses {
		if status != StatusWantToRead && status != StatusReading && status != StatusFinished {
			t.Errorf("status %s should be valid", status)
		}
	}
}
