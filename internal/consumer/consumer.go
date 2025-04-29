package consumer

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dummymq/internal/api"
	"dummymq/internal/model"

	"github.com/go-chi/render"
)

func Run(port int) {
	router := chi.NewRouter()
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Post("/consume", func(w http.ResponseWriter, r *http.Request) {
		body := api.ParseBody[model.Message](w, r)
		fmt.Printf("Consumer on port %v got: %v\r\n", port, body)
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	address := fmt.Sprintf("0.0.0.0:%v", port)

	srv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Cannot start consumer", err)
		}
	}()

	fmt.Printf("Consumer started on %v\r\n", address)

	<-done
	fmt.Println("Stopping consumer")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Failed to stop consumer", err)
		return
	}

	fmt.Println("Consumer stopped")
}
