package dao

import (
	"testing"
)

func TestApplyTemplateInputValidate(t *testing.T) {
	tests := []struct {
		name         string
		input        ApplyTemplateInput
		wantErr      bool
		expectedName string
	}{
		{
			name: "Event is empty",
			input: ApplyTemplateInput{
				Template: *NewTemplate("", "zh-TW", "subject", "bodyPlaint", "bodyHtml"),
			},
			wantErr: true,
		},
		{
			name: "Lang is empty",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "", "subject", "bodyPlaint", "bodyHtml"),
			},
			wantErr: true,
		},
		{
			name: "Subject is empty",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "lang", "", "bodyPlaint", "bodyHtml"),
			},
			wantErr: true,
		},
		{
			name: "Both Body.Plaint and Body.Html are empty",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "lang", "subject", "", ""),
			},
			wantErr: true,
		},
		{
			name: "All fields are valid",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "lang", "subject", "plaint", "html"),
			},
			wantErr:      false,
			expectedName: "event_lang",
		},
		{
			name: "Event, Lang, and Subject are empty",
			input: ApplyTemplateInput{
				Template: *NewTemplate("", "", "", "plaint", "html"),
			},
			wantErr: true,
		},
		{
			name: "Body.Plaint is empty, but Body.Html is not",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "lang", "subject", "", "html"),
			},
			wantErr:      false,
			expectedName: "event_lang",
		},
		{
			name: "Body.Html is empty, but Body.Plaint is not",
			input: ApplyTemplateInput{
				Template: *NewTemplate("event", "lang", "subject", "plaint", ""),
			},
			wantErr:      false,
			expectedName: "event_lang",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.input.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ApplyTemplateInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.expectedName != tt.input.Template.GetName() {
				t.Errorf("ApplyTemplateInput.Validate() name = %v, want %v", tt.input.Template.GetName(), tt.expectedName)
			}
		})
	}
}
