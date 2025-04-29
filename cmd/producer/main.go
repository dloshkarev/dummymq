package main

import (
	"dummymq/internal/api"
	"dummymq/internal/model"
	"fmt"
	"log"
	"math/rand"
)

func main() {
	fmt.Println("Producer started")
	for i := range 100 {
		q := fmt.Sprintf("q%v", rand.Intn(3))
		url := fmt.Sprintf("http://localhost:8082/v1/queues/%v/messages", q)
		fmt.Printf("'click-%v' for %v \r\n", i, url)
		err := api.SendMessage(url, &model.Message{EventType: fmt.Sprintf("click-%v", i)})
		if err != nil {
			log.Fatal(err)
		}
	}
}
