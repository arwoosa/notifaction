package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotifyMsg interface {
	AddTo(to primitive.ObjectID)
	WriteToHeader(c *gin.Context, headerKey string) error
}

type notifyMsg struct {
	event     string
	from      primitive.ObjectID
	to        []primitive.ObjectID
	data      map[string]string
	userTrans UserTransFunc
}

type UserTransFunc func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error)

func NewNotifyMsg(event string, from, to primitive.ObjectID, data map[string]string, userTrans UserTransFunc) (NotifyMsg, error) {
	if userTrans == nil {
		return nil, fmt.Errorf("userTrans is nil")
	}
	return &notifyMsg{
		event:     event,
		from:      from,
		to:        []primitive.ObjectID{to},
		data:      data,
		userTrans: userTrans,
	}, nil
}

func (msg *notifyMsg) AddTo(to primitive.ObjectID) {
	for _, existTo := range msg.to {
		if existTo == to {
			return
		}
	}
	msg.to = append(msg.to, to)
}

func (msg *notifyMsg) WriteToHeader(c *gin.Context, headerKey string) error {
	if msg.userTrans == nil {
		return errors.New("userTrans is nil")
	}
	userToSourceIdMap, err := msg.userTrans(append(msg.to, msg.from))
	if err != nil {
		return fmt.Errorf("ERROR find user source id: %w", err)
	}
	from, ok := userToSourceIdMap[msg.from]
	if !ok {

		return errors.New("not found from user: " + msg.from.Hex())
	}
	notifyTo := make([]string, 0, len(msg.to))
	for _, t := range msg.to {
		tt, ok := userToSourceIdMap[t]
		if !ok {
			continue
		}
		notifyTo = append(notifyTo, tt)
	}
	jsonData, err := marshalHandler(struct {
		Event string            `json:"event"`
		Data  map[string]string `json:"data"`
		From  string            `json:"from"`
		To    []string          `json:"to"`
	}{
		Event: msg.event,
		Data:  msg.data,
		From:  from,
		To:    notifyTo,
	})
	if err != nil {
		return fmt.Errorf("ERROR marshaling notification: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)

	c.Writer.Header().Set(headerKey, encoded)
	return nil
}

var marshalHandler = json.Marshal
