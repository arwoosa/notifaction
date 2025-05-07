package factory

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arwoosa/notifaction/service/mail"
	"github.com/arwoosa/notifaction/service/mail/aws"
	"github.com/arwoosa/notifaction/service/mail/dao"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func NewApiSender() (mail.ApiSender, error) {
	if mockSendor != nil {
		return newMockSender()
	}
	provider := viper.GetString("mail.provider")
	if provider == "aws" {
		return aws.NewApiSender()
	}
	return nil, errors.New("invalid mail provider")
}

type factoryOpt func(*tplImpl)

func WithAllowedDirs(dirs ...string) factoryOpt {
	return func(a *tplImpl) {
		allow := make([]string, len(dirs))
		for i, d := range dirs {
			path, err := filepath.Abs(d)
			if err != nil {
				continue
			}
			allow[i] = path
		}
		a.allowedDirs = allow
	}
}

func NewTemplate(opts ...factoryOpt) (mail.Template, error) {
	if mockTemplate != nil {
		return newMockTemplate()
	}
	tplImpl := &tplImpl{}
	for _, opt := range opts {
		opt(tplImpl)
	}

	if len(tplImpl.allowedDirs) == 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		tplImpl.allowedDirs = []string{homeDir}
	}

	provider := viper.GetString("mail.provider")
	var store mail.TemplateStore
	var err error
	if provider == "aws" {
		store, err = aws.NewTemplateStore()
	}
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("invalid mail provider")
	}
	tplImpl.store = store

	return tplImpl, nil
}

type tplImpl struct {
	store       mail.TemplateStore
	allowedDirs []string
}

func (a *tplImpl) isFileAllowed(file string) bool {
	if len(a.allowedDirs) == 0 {
		return true
	}
	for _, dir := range a.allowedDirs {
		if strings.HasPrefix(file, dir) {
			return true
		}
	}
	return false
}

func (a *tplImpl) Apply(file string) error {
	absFile, err := filepath.Abs(file)
	if err != nil {
		return err
	}
	if !a.isFileAllowed(absFile) {
		return errors.New("file is not in allowed dir: " + strings.Join(a.allowedDirs, ", "))
	}
	// check file exist
	if _, err := os.Stat(absFile); err != nil {
		return fmt.Errorf("file %s does not exist", file)
	}

	// read file
	data, err := os.ReadFile(filepath.Clean(absFile))
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", absFile, err)
	}
	// yaml unmarshal
	var tplDao dao.ApplyTemplateInput
	if err := yaml.Unmarshal(data, &tplDao); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	// validate template dao
	if err := tplDao.Validate(); err != nil {
		return err
	}

	// check template exist
	name := tplDao.GetName()
	exist, err := a.store.IsTemplateExist(name)
	if err != nil {
		return fmt.Errorf("failed to check template exist: %w", err)
	}
	if exist {
		// update template
		return a.store.UpdateTemplate(&tplDao.Template)
	}

	// aws create template
	return a.store.CreateTpl(&tplDao.Template)
}

func (a *tplImpl) Delete(name string) error {
	exist, err := a.store.IsTemplateExist(name)
	if err != nil {
		return fmt.Errorf("failed to check template exist: %w", err)
	}
	if !exist {
		return errors.New("template does not exist")
	}

	return a.store.Delete(name)
}

func (a *tplImpl) List(nextToken string) (*dao.ListTemplateResponse, error) {
	return a.store.List(nextToken)
}

func (a *tplImpl) Detail(name string) (*dao.DetailTemplateResponse, error) {
	return a.store.Detail(name)
}
