package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewNotifyMsg(t *testing.T) {
	tests := []struct {
		name      string
		event     string
		from      primitive.ObjectID
		to        primitive.ObjectID
		data      map[string]string
		userTrans UserTransFunc
		wantErr   bool
	}{
		{
			name:  "valid input",
			event: "test event",
			from:  primitive.NewObjectID(),
			to:    primitive.NewObjectID(),
			data:  map[string]string{"key": "value"},
			userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
				return nil, nil
			},
			wantErr: false,
		},
		{
			name:      "nil userTrans",
			event:     "test event",
			from:      primitive.NewObjectID(),
			to:        primitive.NewObjectID(),
			data:      map[string]string{"key": "value"},
			userTrans: nil,
			wantErr:   true,
		},
		{
			name:  "empty event",
			event: "",
			from:  primitive.NewObjectID(),
			to:    primitive.NewObjectID(),
			data:  map[string]string{"key": "value"},
			userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
				return nil, nil
			},
			wantErr: false,
		},
		{
			name:  "empty data",
			event: "test event",
			from:  primitive.NewObjectID(),
			to:    primitive.NewObjectID(),
			data:  nil,
			userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
				return nil, nil
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNotifyMsg(tt.event, tt.from, tt.to, tt.data, tt.userTrans)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewNotifyMsg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNotifyMsgAddTo(t *testing.T) {
	tests := []struct {
		name     string
		initTo   []primitive.ObjectID
		toAdd    primitive.ObjectID
		expected []primitive.ObjectID
	}{
		{
			name:     "add to empty list",
			initTo:   nil,
			toAdd:    primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
			expected: []primitive.ObjectID{primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
		},
		{
			name:   "add to non-empty list",
			initTo: []primitive.ObjectID{primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
			toAdd:  primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
			expected: []primitive.ObjectID{
				primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
		},
		{
			name:     "add duplicate recipient",
			initTo:   []primitive.ObjectID{[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
			toAdd:    [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expected: []primitive.ObjectID{[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
		},
		{
			name:     "add zero-value recipient",
			initTo:   nil,
			toAdd:    [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expected: []primitive.ObjectID{[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &notifyMsg{to: tt.initTo}
			msg.AddTo(tt.toAdd)
			if !reflect.DeepEqual(msg.to, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, msg.to)
			}
		})
	}
}

func TestWriteToHeader(t *testing.T) {
	tests := []struct {
		name             string
		userTrans        UserTransFunc
		msg              *notifyMsg
		headerKey        string
		wantErr          bool
		wantHeader       string
		wantNotifyTo     []string
		marshalerHandler func(interface{}) ([]byte, error)
	}{
		{
			name: "success",
			msg: &notifyMsg{
				event: "test-event",
				from:  primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				to:    []primitive.ObjectID{primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
				data:  map[string]string{"key": "value"},
				userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
					return map[primitive.ObjectID]string{
						primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "from-source-id",
						primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "to-source-id",
					}, nil
				},
			},
			headerKey:  "X-Notify",
			wantErr:    false,
			wantHeader: `{"event":"test-event","data":{"key":"value"},"from":"from-source-id","to":["to-source-id"]}`,
			wantNotifyTo: []string{
				"to-source-id",
			},
		},
		{
			name: "userTrans nil",
			msg: &notifyMsg{
				event:     "test-event",
				from:      primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				to:        []primitive.ObjectID{primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
				data:      map[string]string{"key": "value"},
				userTrans: nil,
			},
			headerKey: "X-Notify",
			wantErr:   true,
		},
		{
			name: "userTrans error",
			msg: &notifyMsg{
				event: "test-event",
				from:  primitive.NewObjectID(),
				to:    []primitive.ObjectID{primitive.NewObjectID()},
				data:  map[string]string{"key": "value"},
				userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
					return nil, errors.New("userTrans error")
				},
			},
			headerKey: "X-Notify",
			wantErr:   true,
		},
		{
			name: "msg.from not found",

			msg: &notifyMsg{
				event: "test-event",
				from:  primitive.NewObjectID(),
				to:    []primitive.ObjectID{primitive.NewObjectID()},
				data:  map[string]string{"key": "value"},
				userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
					return map[primitive.ObjectID]string{
						primitive.NewObjectID(): "to-source-id",
					}, nil
				},
			},
			headerKey: "X-Notify",
			wantErr:   true,
		},
		{
			name: "notifyTo filtered",
			msg: &notifyMsg{
				event: "test-event",
				from:  primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				to:    []primitive.ObjectID{primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}), primitive.NewObjectID()},
				data:  map[string]string{"key": "value"},
				userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
					return map[primitive.ObjectID]string{
						primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "from-source-id",
						primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "to-source-id",
					}, nil
				},
			},
			headerKey: "X-Notify",
			wantErr:   false,
			wantNotifyTo: []string{
				"to-source-id",
			},
		},
		{
			name: "json marshal error",
			msg: &notifyMsg{
				event: "test-event",
				from:  primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
				to:    []primitive.ObjectID{primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})},
				data:  map[string]string{"key": "value"},
				userTrans: func(userIds []primitive.ObjectID) (map[primitive.ObjectID]string, error) {
					return map[primitive.ObjectID]string{
						primitive.ObjectID([12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "from-source-id",
						primitive.ObjectID([12]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}): "to-source-id",
					}, nil
				},
			},
			wantErr: true,
			marshalerHandler: func(interface{}) ([]byte, error) {
				return nil, errors.New("json marshal error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.marshalerHandler != nil {
				marshalHandler = tt.marshalerHandler
			}
			defer func() {
				marshalHandler = json.Marshal
			}()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			err := tt.msg.WriteToHeader(c, tt.headerKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteToHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				headerValue := c.Writer.Header().Get(tt.headerKey)
				decoded, err := base64.StdEncoding.DecodeString(headerValue)
				if err != nil {
					t.Errorf("failed to decode header value: %v", err)
					return
				}
				var notify struct {
					Event string            `json:"event"`
					Data  map[string]string `json:"data"`
					From  string            `json:"from"`
					To    []string          `json:"to"`
				}
				err = json.Unmarshal(decoded, &notify)
				if err != nil {
					t.Errorf("failed to unmarshal header value: %v", err)
					return
				}
				if notify.Event != tt.msg.event {
					t.Errorf("notify.Event = %q, want %q", notify.Event, tt.msg.event)
				}
				if !reflect.DeepEqual(notify.Data, tt.msg.data) {
					t.Errorf("notify.Data = %v, want %v", notify.Data, tt.msg.data)
				}
				if !reflect.DeepEqual(notify.To, tt.wantNotifyTo) {
					t.Errorf("notify.To = %v, want %v", notify.To, tt.wantNotifyTo)
				}
			}
		})
	}
}
