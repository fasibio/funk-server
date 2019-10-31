package main

import (
	"net/http"
	"net/http/httptest"
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
	if !reflect.DeepEqual(db.wantMessage.Attributes, data.Attributes) {
		db.t.Errorf("DB and want attributes are not equals want %v got %v", db.wantMessage.Attributes, data.Attributes)
	}
	db.tisExecute = true
}

func (db *DBMock) SetIlmPolicy(p DataRolloverPattern) error {
	return nil
}

func (db *DBMock) SetPolicyTemplate() error {
	return nil
}

func TestDataServiceWebSocket_SubscribeTestsCallWithNoWebsocketClient(t *testing.T) {
	d := DataServiceWebSocket{
		ConnectionAllowed: func(r *http.Request) bool {
			return true
		},
		genUID: func() (string, error) { return "1234", nil },
	}
	s := httptest.NewServer(http.HandlerFunc(d.Subscribe))
	defer s.Close()
	res, err := http.Get(s.URL)
	if err != nil {
		t.Fatalf("Error by get request %v", err.Error())
	}
	if res.StatusCode != 400 {
		t.Errorf("Make http request to websocket  want statuscode %v got %v", 400, res.StatusCode)
	}
}

func TestDataServiceWebSocket_SubscribeTests(t *testing.T) {
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
			name:                    "ConnectionAllowed returned allow connection so connection will be open and log message will be send",
			connectionAllowedResult: true,
			want:                    101,
			wantConnectionSize:      1,
			getMessageObjToSend: func() Message {
				return Message{
					Time: mocktime,
					Type: MessageTypeLog,
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
			name:                    "ConnectionAllowed returned allow connection so connection will be open and statsmessage will be send",
			connectionAllowedResult: true,
			want:                    101,
			wantConnectionSize:      1,
			getMessageObjToSend: func() Message {
				return Message{
					Time: mocktime,
					Type: MessageTypeStats,
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

func Test_getIndexDate(t *testing.T) {
	tests := []struct {
		name      string
		time      func() time.Time
		indextype DataRolloverPattern
		want      string
	}{
		{
			name: "is weekly so Returns 2016-5",
			time: func() time.Time {
				res, _ := time.Parse("2006-01-02", "2016-02-02")
				return res
			},
			indextype: Weekly,
			want:      "2016-5",
		},
		{
			name: "is weekly so Returns 2019-44",
			time: func() time.Time {
				res, _ := time.Parse("2006-01-02", "2019-10-31")
				return res
			},
			indextype: Weekly,
			want:      "2019-44",
		},
		{
			name: "is daily so Returns 2019-10-31",
			time: func() time.Time {
				res, _ := time.Parse("2006-01-02", "2019-10-31")
				return res
			},
			indextype: Daily,
			want:      "2019-10-31",
		},
		{
			name: "is monthly so Returns 2019-10",
			time: func() time.Time {
				res, _ := time.Parse("2006-01-02", "2019-10-31")
				return res
			},
			indextype: Monthly,
			want:      "2019-10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getIndexDate(tt.time(), tt.indextype); got != tt.want {
				t.Errorf("getIndexDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataServiceWebSocket_Root(t *testing.T) {
	d := DataServiceWebSocket{}
	s := httptest.NewServer(http.HandlerFunc(d.Root))
	res, err := http.Get(s.URL)
	if err != nil {
		t.Fatalf("Error by get root url: %v", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Want statuscode %v but got %v", 200, res.StatusCode)
	}
}
