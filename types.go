package main

import "time"

// MessageType is the kind of Message which will send
type MessageType string

const (
	// MessageTypeLog a logmessage
	MessageTypeLog MessageType = "LOG"
	//MessageTypeStats a statsmessage
	MessageTypeStats MessageType = "STATS"
)

// Message is the Lawobject between agent and server
// Its the JSON which will be send to server
type Message struct {
	Time          time.Time   `json:"time,omitempty"`        // Time is the explizit time where this dataset is created
	Type          MessageType `json:"type,omitempty"`        // Type is this a LOG or a STATS dataset
	Data          []string    `json:"data,omitempty"`        // Data is an array of seralized JSON. Here are the Jsonobjects from logging Container
	SearchIndex   string      `json:"searchindex,omitempty"` // SearchIndex is the Elasticsearch index to save the given dataset
	Attributes    Attributes  `json:"attr,omitempty"`        // Attributes are Metainformation
	StaticContent string      `json:"static_content,omitempty"`
}

// Attributes are the Metainformation
// like Hostname, the id of tracking container and so on
type Attributes struct {
	Host          string `json:"hostname,omitempty"`
	Containername string `json:"container,omitempty"`
	Servicename   string `json:"service,omitempty"`
	Namespace     string `json:"namespace,omitempty"`
	ContainerID   string `json:"container_id,omitempty"`
}
