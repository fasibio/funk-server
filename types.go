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
	Data        []string    `json:"data,omitempty"`
	SearchIndex string      `json:"searchindex,omitempty"`
	Attributes  Attributes  `json:"attr,omitempty"`
}

type Attributes struct {
	Host          string `json:"hostname,omitempty"`
	Containername string `json:"container,omitempty"`
	Servicename   string `json:"service,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ContainerID   string `json:"container_id,omitempty"`
}
