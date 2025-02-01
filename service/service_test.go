package service

import (
	"testing"
)

func TestGetTemplateName(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		lang     string
		expected string
	}{
		{
			name:     "empty event and lang",
			event:    "",
			lang:     "",
			expected: "_",
		},
		{
			name:     "non-empty event and empty lang",
			event:    "test_event",
			lang:     "",
			expected: "test_event_",
		},
		{
			name:     "empty event and non-empty lang",
			event:    "",
			lang:     "en-US",
			expected: "_en-US",
		},
		{
			name:     "non-empty event and lang",
			event:    "test_event",
			lang:     "en-US",
			expected: "test_event_en-US",
		},
		{
			name:     "event and lang with special characters",
			event:    "test_event!",
			lang:     "en-US@#$",
			expected: "test_event!_en-US@#$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := GetTemplateName(tt.event, tt.lang)
			if actual != tt.expected {
				t.Errorf("GetTemplateName() = %q, want %q", actual, tt.expected)
			}
		})
	}
}

func TestGetTemplateName2(t *testing.T) {
	tests := []struct {
		name     string
		event    string
		lang     string
		expected string
	}{
		{
			name:     "empty event and language",
			event:    "",
			lang:     "",
			expected: "_",
		},
		{
			name:     "non-empty event and empty language",
			event:    "test_event",
			lang:     "",
			expected: "test_event_",
		},
		{
			name:     "empty event and non-empty language",
			event:    "",
			lang:     "en-US",
			expected: "_en-US",
		},
		{
			name:     "non-empty event and language",
			event:    "test_event",
			lang:     "en-US",
			expected: "test_event_en-US",
		},
		{
			name:     "event and language with special characters",
			event:    "test_event!",
			lang:     "en-US@#$",
			expected: "test_event!_en-US@#$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Notification{
				Event: tt.event,
				Lang:  tt.lang,
			}
			actual := n.GetTemplateName()
			if actual != tt.expected {
				t.Errorf("GetTemplateName() = %q, want %q", actual, tt.expected)
			}
		})
	}
}
