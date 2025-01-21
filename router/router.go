package router

import (
	"github.com/94peter/microservice/apitool"
)

func GetApis() []apitool.GinAPI {
	return []apitool.GinAPI{
		&notification{},
	}
}
