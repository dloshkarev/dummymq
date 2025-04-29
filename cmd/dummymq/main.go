package main

import (
	"context"
	"dummymq/internal/api"
	"dummymq/internal/config"
	"dummymq/internal/engine"
	. "dummymq/internal/engine"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func main() {
	fmt.Println("Starting dummymq")
	cfg := config.MustLoad()

	mq := engine.Run(cfg.MQEngine.Queues, cfg.MQEngine.MessageLimit)

	createAndRunApiServer(cfg, mq)
}

func createAndRunApiServer(cfg *config.Config, engine *Engine) {
	router := chi.NewRouter()
	router.Use(render.SetContentType(render.ContentTypeJSON))

	router.Route("/v1/queues/{queue_name}", func(r chi.Router) {
		r.Post("/subscriptions", api.Subscribe(engine, cfg.MQEngine.MessageLimit))
		r.Post("/messages", api.AddMessage(engine))
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Cannot start server", err)
		}
	}()

	fmt.Printf("Server started on %v\r\n", cfg.HTTPServer.Address)

	<-done
	fmt.Println("Stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Failed to stop server", err)
		return
	}

	fmt.Println("Server stopped")
}
