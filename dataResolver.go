package main

import (
	"encoding/json"
	"log"
	"net/http"

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
		log.Println("Connection forbidden")
		w.WriteHeader(401)
	}
	log.Println("New Subscribe Client")
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("error by upgrade", err)
		return
	}
	uuid, _ := u.genUID()

	go func() {
		for {
			var messages []Message
			err := c.ReadJSON(&messages)

			if err != nil {
				log.Println("error by ClientConn", err)
				delete(u.ClientConnections, uuid)
				return
			}

			for _, msg := range messages {
				str := msg.Data
				var d interface{}

				for _, v := range str {
					err = json.Unmarshal([]byte(v), &d)
					if err != nil {
						log.Println(err)
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
						}, msg.SearchIndex)

					case MessageType_Stats:
						{
							u.Db.AddStats(StatsData{
								Containername: msg.Containername,
								Timestamp:     msg.Time,
								Host:          msg.Host,
								Type:          string(msg.Type),
								Stats:         d,
								ContainerID:   msg.ContainerID,
							}, msg.SearchIndex)
						}
					}
				}
			}
		}
	}()

	u.ClientConnections[uuid] = c

}
