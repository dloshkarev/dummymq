package model

type Message struct {
	EventType string `json:"event_type" validate:"required,string"`
}

type Subscriber struct {
	Endpoint string `json:"endpoint" validate:"required,url"`
}
