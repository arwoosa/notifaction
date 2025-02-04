package router

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/94peter/microservice/apitool"
	"github.com/94peter/microservice/apitool/err"
	"github.com/gin-gonic/gin"
)

type test struct {
	err.CommonErrorHandler
}

func (m *test) GetHandlers() []*apitool.GinHandler {
	return []*apitool.GinHandler{
		{
			Path:    "/test/header2post",
			Method:  "POST",
			Handler: m.testHeader2Post,
		},
	}
}

func (m *test) testHeader2Post(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		m.GinErrorWithStatusHandler(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	headerValue := base64.StdEncoding.EncodeToString(body)
	c.Writer.Header().Set("X-Notify", headerValue)
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}
