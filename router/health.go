package router

import (
	"fmt"

	"github.com/94peter/microservice/apitool"
	"github.com/94peter/microservice/apitool/err"
	"github.com/arwoosa/notifaction/service/identity"
	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/gin-gonic/gin"
)

type health struct {
	err.CommonErrorHandler
	isIdentityServiceReady bool
	isEmailServiceReady    bool
}

func (m *health) GetHandlers() []*apitool.GinHandler {
	return []*apitool.GinHandler{
		{
			Path:    "/health/alive",
			Method:  "GET",
			Handler: aliveHandler,
		},
		{
			Path:    "/health/ready",
			Method:  "GET",
			Handler: m.readyHandler,
		},
	}
}

func aliveHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "I am alive",
	})
}

func (m *health) readyHandler(c *gin.Context) {
	if !m.isIdentityServiceReady {
		serice, err := identity.NewIdentity()
		if err != nil {
			m.GinErrorHandler(c, fmt.Errorf("identity service is not ready: %w", err))
			return
		}
		ready, err := serice.IsReady()
		if err != nil {
			m.GinErrorHandler(c, fmt.Errorf("identity service is not ready: %w", err))
			return
		}
		if !ready {
			m.GinErrorHandler(c, fmt.Errorf("identity service is not ready"))
			return
		}
		m.isIdentityServiceReady = true
	}
	if !m.isEmailServiceReady {
		serice, err := factory.NewTemplate()
		if err != nil {
			m.GinErrorHandler(c, fmt.Errorf("email service is not ready: %w", err))
			return
		}
		_, err = serice.List("")
		if err != nil {
			m.GinErrorHandler(c, fmt.Errorf("email service is not ready: %w", err))
			return
		}
		m.isEmailServiceReady = true
	}
	c.JSON(200, gin.H{
		"message": "service notification is ready",
	})
	return
}
