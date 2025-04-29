package main

import (
	"dummymq/internal/consumer"
	"os"
	"strconv"
)

func main() {
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	consumer.Run(port)
}
