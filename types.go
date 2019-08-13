package main

import "time"

type MessageType string

const (
	MessageType_Log   MessageType = "LOG"
	MessageType_Stats MessageType = "STATS"
)

type Message struct {
	Time        time.Time   `json:"time,omitempty"`
	Type        MessageType `json:"type,omitempty"`
	Data        []string    `json:"data,omitempty"`        //a list of Stringjson with the explizit data
	SearchIndex string      `json:"searchindex,omitempty"` //the elasticsearch index to set this data
	Attributes  Attributes  `json:"attr,omitempty"`        // Meta information
}

type Attributes struct {
	Host          string `json:"hostname,omitempty"`
	Containername string `json:"container,omitempty"`
	Servicename   string `json:"service,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ContainerID   string `json:"container_id,omitempty"`
}
