package main

import "time"

type MessageType string

const (
	MessageType_Log   MessageType = "LOG"
	MessageType_Stats MessageType = "STATS"
)

type Message struct {
	Time          time.Time
	Type          MessageType
	Data          []string
	Containername string
	ContainerID   string
	Host          string
	SearchIndex   string
}
