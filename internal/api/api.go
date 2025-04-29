package api

import (
	"bytes"
	"dummymq/internal/engine"
	. "dummymq/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/render"
)

func AddMessage(engine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueName := r.PathValue("queue_name")
		body := ParseBody[Message](w, r)
		err := engine.AddMessage(queueName, &body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Cannot send message to %v, req = %v", queueName, body), 400)
		}
		return
	}
}

func Subscribe(engine *engine.Engine, msgLimit int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		queueName := r.PathValue("queue_name")
		body := ParseBody[Subscriber](w, r)

		ch, err := engine.Subscribe(queueName, body.Endpoint, msgLimit)
		if err == nil {
			fmt.Printf("Endpoint %v listening %v\r\n", body.Endpoint, queueName)
			go func(endpoint string) {
				for {
					select {
					case msg := <-ch:
						err := SendMessage(endpoint, msg)
						if err != nil {
							msg := fmt.Sprintf("Cannot send message to endpoint = %v, req = %v, err = %v", endpoint, body, err)
							fmt.Println(msg)
						}
					default:
						// avoid cpu burn
						time.Sleep(100 * time.Millisecond)
					}
				}
			}(body.Endpoint)
		} else {
			msg := fmt.Sprintf("Cannot subsribe to %v, req = %v, err = %v", queueName, body, err)
			fmt.Println(msg)
			http.Error(w, msg, 400)
		}
	}
}

func ParseBody[T any](w http.ResponseWriter, r *http.Request) T {
	var req T

	err := render.DecodeJSON(r.Body, &req)
	if errors.Is(err, io.EOF) {
		msg := "Request body is empty"
		fmt.Println(msg)
		http.Error(w, msg, 400)
		return req
	}
	if err != nil {
		msg := "Failed to decode request"
		http.Error(w, msg, 400)
		fmt.Println(msg)
		return req
	}

	//fmt.Printf("Request body decoded: %v\r\n", req)
	return req
}

func SendMessage(url string, message *Message) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			fmt.Println("failed to close response body")
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send successful request. Status was %q", response.Status)
	}

	return nil
}
