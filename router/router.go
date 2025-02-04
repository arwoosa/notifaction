package router

import (
	"github.com/94peter/microservice/apitool"
	"github.com/spf13/viper"
)

func GetApis() []apitool.GinAPI {
	apis := []apitool.GinAPI{
		&notification{},
		&health{},
	}
	if viper.GetBool("api.test") {
		apis = append(apis, &test{})
	}
	return apis
}
