package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/94peter/microservice/apitool"
	"github.com/94peter/microservice/apitool/err"
	"github.com/arwoosa/notifaction/router/request"
	"github.com/arwoosa/notifaction/service"
	"github.com/arwoosa/notifaction/service/identity"
	"github.com/arwoosa/notifaction/service/mail/factory"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type notification struct {
	err.CommonErrorHandler
}

var header2data []string

func newNotification() *notification {
	header2data = viper.GetStringSlice("mail.header2data")
	return &notification{}
}

func (m *notification) GetHandlers() []*apitool.GinHandler {
	return []*apitool.GinHandler{
		{
			Path:    "/notification",
			Method:  "POST",
			Handler: m.createNotification,
		},
	}
}

// https://app.apidog.com/link/project/607604/apis/api-13258249?branchId=767188
func (m *notification) createNotification(c *gin.Context) {
	var requestBody request.CreateNotification
	if err := c.BindJSON(&requestBody); err != nil {
		m.GinErrorWithStatusHandler(c, http.StatusBadRequest, fmt.Errorf("invalid request body: %w", err))
		return
	}
	if err := requestBody.Validate(); err != nil {
		m.GinErrorWithStatusHandler(c, http.StatusBadRequest, err)
		return
	}
	sender, err := factory.NewApiSender()
	if err != nil {
		m.GinErrorHandler(c, err)
		return
	}

	classificationLang, err := identity.NewIdentity()
	if err != nil {
		m.GinErrorHandler(c, err)
		return
	}
	cl, err := classificationLang.SubToInfo(requestBody.From, requestBody.To)
	if err != nil {
		m.GinErrorHandler(c, err)
		return
	}
	requestBody.Data["FROM"] = cl.From.Name
	for _, h := range header2data {
		if c.Request.Header.Get(h) == "" {
			requestBody.Data[h] = "missing header: " + h
			continue
		}
		requestBody.Data[h] = c.Request.Header.Get(h)
	}
	var errSends []sendError
	successResp := make([]gin.H, 0)
	for _, lang := range cl.GetLangs() {
		for i, info := range cl.GetInfos(lang) {
			if i > 0 {
				time.Sleep(200 * time.Microsecond)
			}
			requestBody.Data["TO"] = info.Name
			mid, err := sender.Send(&service.Notification{
				Event:  requestBody.Event,
				Lang:   lang,
				From:   cl.From,
				SendTo: []*service.Info{info},
				Data:   requestBody.Data,
			})
			if err != nil {
				errSends = append(errSends, sendError{err: err, info: info})
			} else {
				successResp = append(successResp, gin.H{
					"send_to": info.Name,
					"mid":     mid,
					"lang":    lang,
					"from":    cl.From.Name,
					"event":   requestBody.Event,
				})
			}
		}
	}

	if len(errSends) == 0 {
		c.JSON(http.StatusAccepted, gin.H{
			"success": successResp,
		})
		return
	}

	if len(errSends) == len(requestBody.To) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": errSends[0].err.Error(),
		})
		return
	}

	errorResp := make([]gin.H, len(errSends))
	for i, e := range errSends {
		errorResp[i] = e.output()
	}
	c.JSON(http.StatusPartialContent, errorResp)
}

type sendError struct {
	err  error
	info *service.Info
}

func (s *sendError) output() gin.H {
	return gin.H{
		"error": s.err.Error(),
		"email": s.info.Name,
	}
}
