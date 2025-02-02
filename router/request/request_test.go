package request

import (
	"testing"
)

func TestCreateNotificationValidate(t *testing.T) {
	tests := []struct {
		name    string
		notify  *CreateNotification
		wantErr bool
	}{
		{
			name: "empty to",
			notify: &CreateNotification{
				To:    []string{},
				From:  "test",
				Event: "test",
				Data:  map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "empty from",
			notify: &CreateNotification{
				To:    []string{"test"},
				From:  "",
				Event: "test",
				Data:  map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "empty event",
			notify: &CreateNotification{
				To:    []string{"test"},
				From:  "test",
				Event: "",
				Data:  map[string]string{},
			},
			wantErr: true,
		},
		{
			name: "empty data",
			notify: &CreateNotification{
				To:    []string{"test"},
				From:  "test",
				Event: "test",
				Data:  nil,
			},
			wantErr: true,
		},
		{
			name: "valid notification",
			notify: &CreateNotification{
				To:    []string{"test"},
				From:  "test",
				Event: "test",
				Data:  map[string]string{},
			},
			wantErr: false,
		},
		{
			name:    "nil notification",
			notify:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.notify.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
