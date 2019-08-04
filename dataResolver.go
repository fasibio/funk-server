package main

import (
	"encoding/json"
	"net/http"

	"github.com/fasibio/funk-server/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

type DataServiceWebSocket struct {
	ClientConnections map[string]*websocket.Conn
	genUID            func() (string, error)
	Db                *KonfigData
	ConnectionAllowed func(*http.Request) bool
}

func (u *DataServiceWebSocket) Root(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hallo vom Server"))
}

func (u *DataServiceWebSocket) Subscribe(w http.ResponseWriter, r *http.Request) {
	if !u.ConnectionAllowed(r) {
		logger.Get().Infow("Connection forbidden to subscribe")
		w.WriteHeader(401)
	}
	uuid, _ := u.genUID()
	logger.Get().Infow("New Subscribe Client", "subscriptionID", uuid)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Get().Errorw("error by Subscribe"+err.Error(), "subscriptionID", uuid)
		return
	}

	go func() {
		for {
			var messages []Message
			err := c.ReadJSON(&messages)

			if err != nil {
				logger.Get().Errorw("error by ClientConn"+err.Error(), "subscriptionID", uuid)
				delete(u.ClientConnections, uuid)
				return
			}

			for _, msg := range messages {
				str := msg.Data
				var d interface{}

				for _, v := range str {
					err = json.Unmarshal([]byte(v), &d)
					if err != nil {
						logger.Get().Errorw("error by unmarshal data:"+err.Error(), "subscriptionID", uuid)
						d = v
					}

					switch msg.Type {
					case MessageType_Log:
						u.Db.AddLog(LogData{
							Containername: msg.Containername,
							Timestamp:     msg.Time,
							Host:          msg.Host,
							Type:          string(msg.Type),
							Logs:          d,
							ContainerID:   msg.ContainerID,
						}, msg.SearchIndex+"_funk")

					case MessageType_Stats:
						{
							u.Db.AddStats(StatsData{
								Containername: msg.Containername,
								Timestamp:     msg.Time,
								Host:          msg.Host,
								Type:          string(msg.Type),
								Stats:         d,
								ContainerID:   msg.ContainerID,
							}, msg.SearchIndex+"_funk")
						}
					}
				}
			}
		}
	}()

	u.ClientConnections[uuid] = c

}
