package factory

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"
)

func TestNewApiSender(t *testing.T) {
	viper.Set("aws.ses.region", "region")
	viper.Set("aws.ses.credentails.filename", "../aws/test_credentials")
	viper.Set("aws.ses.credentails.profile", "default")
	viper.Set("mail.from", "test@example.com")
	tests := []struct {
		name       string
		provider   string
		wantErr    bool
		wantSender bool
	}{
		{
			name:       "AWS provider",
			provider:   "aws",
			wantErr:    false,
			wantSender: true,
		},
		{
			name:       "Non-AWS provider",
			provider:   "other",
			wantErr:    true,
			wantSender: false,
		},
		{
			name:       "Empty provider",
			provider:   "",
			wantErr:    true,
			wantSender: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("mail.provider", tt.provider)
			sender, err := NewApiSender()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApiSender() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantSender && sender == nil {
				t.Errorf("NewApiSender() sender = %v, want %v", sender, tt.wantSender)
			}
		})
	}
}

func TestNewTemplate(t *testing.T) {
	wantDir, _ := os.UserHomeDir()
	projectPath, _ := filepath.Abs("../../../")
	tests := []struct {
		name     string
		preFunc  func()
		opts     []factoryOpt
		wantErr  bool
		wantDirs []string
	}{
		{
			name: "no options",
			preFunc: func() {
				viper.Set("aws.ses.region", "region")
				viper.Set("aws.ses.credentails.filename", "../aws/test_credentials")
				viper.Set("aws.ses.credentails.profile", "default")
				viper.Set("mail.from", "test@example.com")
				viper.Set("mail.provider", "aws")
			},
			wantDirs: []string{wantDir},
		},
		{
			name: "with allowed directories option",
			opts: []factoryOpt{WithAllowedDirs("../")},
			preFunc: func() {
				viper.Set("aws.ses.region", "region")
				viper.Set("aws.ses.credentails.filename", "../aws/test_credentials")
				viper.Set("aws.ses.credentails.profile", "default")
				viper.Set("mail.from", "test@example.com")
				viper.Set("mail.provider", "aws")
			},
			wantDirs: []string{projectPath + "/service/mail"},
		},
		{
			name:     "invalid mail provider",
			preFunc:  func() { viper.Set("mail.provider", "notexist") },
			wantErr:  true,
			wantDirs: nil,
		},
		{
			name:     "new template with aws provider fail",
			preFunc:  func() { viper.Set("mail.provider", "aws") },
			wantDirs: []string{wantDir},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()
			if tt.preFunc != nil {
				tt.preFunc()
			}
			tpl, err := NewTemplate(tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tpl != nil {
				t.Errorf("NewTemplate() tpl = %v, want %v", tpl, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if tpl == nil {
				t.Errorf("NewTemplate() tpl = %v, want %v", tpl, tt.wantErr)
			}
			if !slices.Equal(tt.wantDirs, tpl.(*tplImpl).allowedDirs) {
				t.Errorf("NewTemplate() allowedDirs = %v, want %v", tpl.(*tplImpl).allowedDirs, tt.wantDirs)
			}
		})

	}

}

func TestTplImpl_isFileAllowed(t *testing.T) {
	tests := []struct {
		name        string
		allowedDirs []string
		file        string
		wantAllowed bool
	}{
		{
			name:        "empty allowed directories",
			allowedDirs: []string{},
			file:        "/path/to/file",
			wantAllowed: true,
		},
		{
			name:        "file path matches one of the allowed directories",
			allowedDirs: []string{"/path/to"},
			file:        "/path/to/file",
			wantAllowed: true,
		},
		{
			name:        "file path does not match any of the allowed directories",
			allowedDirs: []string{"/path/to"},
			file:        "/other/path/to/file",
			wantAllowed: false,
		},
		{
			name:        "file path is a subdirectory of one of the allowed directories",
			allowedDirs: []string{"/path/to"},
			file:        "/path/to/subdir/file",
			wantAllowed: true,
		},
		{
			name:        "file path is an absolute path",
			allowedDirs: []string{"/"},
			file:        "/path/to/file",
			wantAllowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tplImpl{allowedDirs: tt.allowedDirs}
			if got := a.isFileAllowed(tt.file); got != tt.wantAllowed {
				t.Errorf("isFileAllowed() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}

func TestTplImpl_Delete(t *testing.T) {
	tests := []struct {
		name         string
		mailTemplate mail.TemplateStore
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name: "template exists and deletion is successful",
			mailTemplate: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return true, nil
				}),
				mail.WithDeleteTemplate(func(name string) error {
					return nil
				}),
			),
			wantErr: false,
		},
		{
			name: "template does not exist",
			mailTemplate: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return false, nil
				}),
				mail.WithDeleteTemplate(func(name string) error {
					return nil
				}),
			),
			wantErr:    true,
			wantErrMsg: "template does not exist",
		},
		{
			name: "error occurs while checking template existence",
			mailTemplate: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return false, errors.New("mock error")
				}),
				mail.WithDeleteTemplate(func(name string) error {
					return nil
				}),
			),
			wantErr:    true,
			wantErrMsg: "failed to check template exist: mock error",
		},
		{
			name: "error occurs during deletion",
			mailTemplate: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return true, nil
				}),
				mail.WithDeleteTemplate(func(name string) error {
					return errors.New("mock error")
				}),
			),
			wantErr:    true,
			wantErrMsg: "mock error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tplImpl := &tplImpl{store: tt.mailTemplate}

			err := tplImpl.Delete("test-template")
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrMsg != "" && err.Error() != tt.wantErrMsg {
				t.Errorf("Delete() error message = %v, want %v", err.Error(), tt.wantErrMsg)
			}
		})
	}
}

func TestTplImpl_Apply(t *testing.T) {
	userDir, _ := os.UserHomeDir()
	tests := []struct {
		name       string
		file       string
		allowedDir string
		store      mail.TemplateStore
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "File not in allowed directory",
			file:       "/path/to/disallowed/file.yaml",
			allowedDir: "/path/to/allowed/dir",
			store:      mail.NewMockTemplateStore(),
			wantErr:    true,
			errMsg:     "file is not in allowed dir: /path/to/allowed/dir",
		},
		{
			name:       "File does not exist",
			file:       "/path/to/nonexistent/file.yaml",
			allowedDir: "/path/to",
			store:      mail.NewMockTemplateStore(),
			wantErr:    true,
			errMsg:     "file /path/to/nonexistent/file.yaml does not exist",
		},
		{
			name:       "YAML unmarshal error",
			file:       "./test_invalid_yaml_apply.yaml",
			allowedDir: userDir,
			store:      mail.NewMockTemplateStore(),
			wantErr:    true,
			errMsg:     "failed to unmarshal yaml: yaml: unmarshal error",
		},
		{
			name:       "Template validation error",
			file:       "./test_validation_error.yaml",
			allowedDir: userDir,
			store:      mail.NewMockTemplateStore(),
			wantErr:    true,
			errMsg:     "lang is required",
		},
		{
			name:       "Template exists and update succeeds",
			file:       "./test_valid.yaml",
			allowedDir: userDir,
			store: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return true, nil
				}),
				mail.WithUpdateTemplate(func(tpl *dao.Template) error {
					return nil
				}),
			),
			wantErr: false,
		},
		{
			name:       "Template check exists error",
			file:       "./test_valid.yaml",
			allowedDir: userDir,
			store: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return false, errors.New("check error")
				}),
			),
			wantErr: true,
			errMsg:  "failed to check template exist: check error",
		},
		{
			name:       "Template exists and update fails",
			file:       "./test_valid.yaml",
			allowedDir: userDir,
			store: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return true, nil
				}),
				mail.WithUpdateTemplate(func(tpl *dao.Template) error {
					return errors.New("update error")
				}),
			),
			wantErr: true,
			errMsg:  "update error",
		},
		{
			name:       "Template does not exist and create succeeds",
			file:       "./test_valid.yaml",
			allowedDir: userDir,
			store: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return false, nil
				}),
				mail.WithCreateTemplate(func(tpl *dao.Template) error {
					return nil
				}),
			),
			wantErr: false,
		},
		{
			name:       "Template does not exist and create fails",
			file:       "./test_valid.yaml",
			allowedDir: userDir,
			store: mail.NewMockTemplateStore(
				mail.WithIsTemplateExist(func(name string) (bool, error) {
					return false, nil
				}),
				mail.WithCreateTemplate(func(tpl *dao.Template) error {
					return errors.New("create error")
				}),
			),
			wantErr: true,
			errMsg:  "create error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &tplImpl{allowedDirs: []string{tt.allowedDir}, store: tt.store}
			file, err := filepath.Abs(tt.file)
			if err != nil {
				t.Errorf("Apply() error = %v", err)
				return
			}
			err = a.Apply(file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Apply() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTplImpl_List(t *testing.T) {
	tests := []struct {
		name       string
		nextToken  string
		mockStore  mail.TemplateStore
		storeError error
		wantErr    bool
	}{
		{
			name:      "Test List with empty nextToken",
			nextToken: "",
			mockStore: mail.NewMockTemplateStore(
				mail.WithListTemplate(func(nextToken string) (*dao.ListTemplateResponse, error) {
					emptyStr := ""
					return &dao.ListTemplateResponse{
						NextToken: &emptyStr,
					}, nil
				}),
			),
			storeError: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tplImpl := &tplImpl{
				store: tt.mockStore,
			}

			_, err := tplImpl.List(tt.nextToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("tplImpl.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type mockTemplateStore struct {
	ListFunc func(nextToken string) (*dao.ListTemplateResponse, error)
}

func (m *mockTemplateStore) List(nextToken string) (*dao.ListTemplateResponse, error) {
	return m.ListFunc(nextToken)
}
