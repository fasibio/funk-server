package main

import (
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/posener/wstest"
)

type DBMock struct {
	t           *testing.T
	wantMessage Message
	tisExecute  bool
}

func (db *DBMock) AddLog(data LogData, index string) {
	if !reflect.DeepEqual(db.wantMessage.Attributes, data.Attributes) {
		db.t.Errorf("DB and want attributes are not equals want %v got %v", db.wantMessage.Attributes, data.Attributes)
	}
	db.tisExecute = true
}

func (db *DBMock) AddStats(data StatsData, index string) {

}
func TestDataServiceWebSocket_SubscribeTestThatAbortByConnectionNotAllowed(t *testing.T) {
	mocktime, err := time.Parse("2006-01-02", "2019-06-02")
	if err != nil {
		t.Fatal(err)
	}
	type test struct {
		name                    string
		connectionAllowedResult bool
		want                    int
		wantConnectionSize      int
		getMessageObjToSend     func() Message
		breakAfterStauscheck    bool
	}

	data := []test{
		test{
			name:                    "ConnectionAllowed returned allow connection so connection will be open",
			connectionAllowedResult: true,
			want:                    101,
			wantConnectionSize:      1,
			getMessageObjToSend: func() Message {
				return Message{
					Time: mocktime,
					Type: MessageType_Log,
					Data: []string{
						"{\"test\":\"test\"",
					},
					SearchIndex: "testindex",
					Attributes: Attributes{
						Host: "local.test",
					},
				}
			},
			breakAfterStauscheck: false,
		},
		test{
			name:                    "ConnectionAllowed returned disallow connection so no connection will be open ",
			connectionAllowedResult: false,
			want:                    401,
			wantConnectionSize:      0,
			getMessageObjToSend:     func() Message { return Message{} },
			breakAfterStauscheck:    true,
		},
	}

	const mockHeaderKey = "connectionAllowedResult"

	for _, one := range data {
		dbMock := &DBMock{
			t:           t,
			wantMessage: one.getMessageObjToSend(),
			tisExecute:  false,
		}
		dataserverholder := DataServiceWebSocket{
			Db: dbMock,
			ConnectionAllowed: func(r *http.Request) bool {
				res, _ := strconv.ParseBool(r.Header.Get(mockHeaderKey))
				return res
			},
			ClientConnections: make(map[string]*websocket.Conn),
			genUID:            func() (string, error) { return "1234", nil },
		}
		d := wstest.NewDialer(http.HandlerFunc(dataserverholder.Subscribe))
		httpHeader := make(http.Header)
		httpHeader.Add(mockHeaderKey, strconv.FormatBool(one.connectionAllowedResult))
		c, resp, _ := d.Dial("ws://"+"whatever"+"/", httpHeader)
		if one.want != resp.StatusCode {
			t.Errorf("Error by Status want %v got %v", one.want, resp.StatusCode)
		}
		if one.breakAfterStauscheck {
			break
		}

		err := c.WriteJSON([]Message{one.getMessageObjToSend()})
		if err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second) // find a better solution
		if !dbMock.tisExecute {
			t.Errorf("Mock db was not executed")
		}
		if len(dataserverholder.ClientConnections) != one.wantConnectionSize {
			t.Errorf("Error by ClientConnections want %v got %v connections", one.wantConnectionSize, len(dataserverholder.ClientConnections))
		}
	}
}
