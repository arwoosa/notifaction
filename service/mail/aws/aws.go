package aws

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/viper"
)

func newAwsSession() (*session.Session, error) {
	region := viper.GetString("aws.ses.region")
	if region == "" {
		return nil, errors.New("aws.ses.region is empty")
	}
	filename := viper.GetString("aws.ses.credentails.filename")
	if filename == "" {
		return nil, errors.New("aws.ses.credentails.filename is empty")
	}
	profile := viper.GetString("aws.ses.credentails.profile")
	if profile == "" {
		return nil, errors.New("aws.ses.credentails.profile is empty")
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials(filename, profile),
	})
	if err != nil {
		return nil, fmt.Errorf("new aws session fail: %w", err)
	}

	_, err = sess.Config.Credentials.Get()
	if err != nil {
		return nil, fmt.Errorf("get aws credentials fail: %w", err)
	}
	return sess, nil
}
