package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ezhdanovskiy/go-ping-pong/pkg/ponger"
	"github.com/ezhdanovskiy/go-ping-pong/pkg/rabbit"
	"github.com/pkg/errors"
)

func main() {
	url := "amqp://guest:guest@localhost:5672"

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		s := <-stop
		log.Printf("got %q signal", s)
		cancel()
	}()

	client, err := rabbit.NewClient(url)
	if err != nil {
		log.Fatalf("Failed to create AMQP client: %s", err)
	}
	defer client.Close()

	srv := ponger.Service{
		Broker: client,
	}

	n, err := srv.Run(ctx)
	if err != nil && errors.Cause(err) != context.Canceled {
		log.Print("ponger return error:", err)
	}

	log.Printf("%v pings were ponged", n)
}
