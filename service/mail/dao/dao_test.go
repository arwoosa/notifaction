package dao

import (
	"testing"
)

func TestApplyTemplateInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   *ApplyTemplateInput
		wantErr bool
	}{
		{
			name: "Test Validate with empty fields",
			input: &ApplyTemplateInput{
				Template: *NewTemplate("", "", "", "", ""),
			},
			wantErr: true, // TODO: currently Validate always returns nil, so this test will fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.input.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ApplyTemplateInput.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
