package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fasibio/funk-server/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{} // use default options

type DataServiceWebSocket struct {
	ClientConnections map[string]*websocket.Conn
	genUID            func() (string, error)
	Db                ElsticConnection
	ConnectionAllowed func(*http.Request) bool
	rollOverPattern   DataRolloverPattern
}

type Resolver interface {
	Root(w http.ResponseWriter, r *http.Request)
	Subscribe(w http.ResponseWriter, r *http.Request)
}

func (u *DataServiceWebSocket) Root(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("Hallo vom Server"))
	if err != nil {
		logger.Get().Errorw("Error by Root Handler" + err.Error())
	}
}

type DataRolloverPattern string

func (d DataRolloverPattern) GetDeletePhaseString() string {
	switch d {
	case Daily:
		return "30d"
	case Weekly:
		return "90d"
	case Monthly:
		return "180d"
	default:
		//if does not match is weekly
		return "8d"
	}
}

func (d DataRolloverPattern) GetWarmPhaseString() string {
	switch d {
	case Daily:
		return "2d"
	case Weekly:
		return "8d"
	case Monthly:
		return "33d"
	default:
		//if does not match is weekly
		return "8d"
	}
}

func (d DataRolloverPattern) GetColdPhaseString() string {
	switch d {
	case Daily:
		return "15d"
	case Weekly:
		return "30d"
	case Monthly:
		return "60d"
	default:
		//if does not match is weekly
		return "30d"
	}
}

const Daily DataRolloverPattern = "Daily"
const Weekly DataRolloverPattern = "Weekly"
const Monthly DataRolloverPattern = "Monthly"

func getIndexDate(time time.Time, indextype DataRolloverPattern) string {
	switch indextype {
	case Daily:
		return time.Format("2006-01-02")
	case Weekly:
		year, week := time.ISOWeek()
		return fmt.Sprintf("%d-%d", year, week)
	case Monthly:
		return time.Format("2006-01")
	default:
		//if does not match is weekly
		year, week := time.ISOWeek()
		return fmt.Sprintf("%d-%d", year, week)
	}
}

func getLoggerWithSubscriptionID(logs *zap.SugaredLogger, uuid string) *zap.SugaredLogger {
	return logs.With(
		"subscriptionID", uuid,
	)
}

func (u *DataServiceWebSocket) interpretMessage(messages []Message, logs *zap.SugaredLogger) {
	for _, msg := range messages {
		str := msg.Data
		var d interface{}
		var staticContent interface{}
		err := json.Unmarshal([]byte(msg.StaticContent), &staticContent)
		if err != nil {
			logs.Errorw("error by unmarshal staticcontent:"+err.Error(), "staticcontent", msg.StaticContent)
			staticContent = msg.StaticContent
		}
		for _, v := range str {
			err := json.Unmarshal([]byte(v), &d)
			if err != nil {
				logs.Errorw("error by unmarshal data:" + err.Error())
				d = v
			}
			switch msg.Type {
			case MessageTypeLog:
				u.Db.AddLog(LogData{
					Timestamp:     msg.Time,
					Type:          string(msg.Type),
					Logs:          d,
					Attributes:    msg.Attributes,
					StaticContent: staticContent,
				}, msg.SearchIndex+"_funk-"+getIndexDate(time.Now(), u.rollOverPattern))
			case MessageTypeStats:
				{
					u.Db.AddStats(StatsData{
						Timestamp:     msg.Time,
						Type:          string(msg.Type),
						Stats:         d,
						Attributes:    msg.Attributes,
						StaticContent: staticContent,
					}, getStatsIndexName(msg.SearchIndex)+getIndexDate(time.Now(), u.rollOverPattern))
				}
			default:
				{
					u.Db.AddStats(StatsData{
						Timestamp:     msg.Time,
						Type:          string(msg.Type),
						Stats:         d,
						Attributes:    msg.Attributes,
						StaticContent: staticContent,
					}, msg.SearchIndex+"_funk-"+getIndexDate(time.Now(), u.rollOverPattern))
				}
			}
		}
	}
}

func getStatsIndexName(name string) string {
	if strings.Contains(name, "_stats_cumulated") {
		return "stats_cumulated_funk-"
	} else {
		return "stats_funk-"
	}

}

func (u *DataServiceWebSocket) messageSubscribeHandler(uuid string, c *websocket.Conn) {
	logs := getLoggerWithSubscriptionID(logger.Get(), uuid)
	for {
		var messages []Message
		err := c.ReadJSON(&messages)
		if err != nil {
			logs.Errorw("error by ClientConn" + err.Error())
			delete(u.ClientConnections, uuid)
			return
		}

		u.interpretMessage(messages, logs)
	}
}

func (u *DataServiceWebSocket) Subscribe(w http.ResponseWriter, r *http.Request) {
	if !u.ConnectionAllowed(r) {
		logger.Get().Infow("Connection forbidden to subscribe", "ForwardedIP", r.Header.Get("X-Forwarded-For"))
		w.WriteHeader(401)
		return
	}
	uuid, _ := u.genUID()
	logs := getLoggerWithSubscriptionID(logger.Get(), uuid)
	logs.Infow("New Subscribe Client")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logs.Errorw("error by Subscribe" + err.Error())
		return
	}

	go u.messageSubscribeHandler(uuid, c)

	u.ClientConnections[uuid] = c
}
