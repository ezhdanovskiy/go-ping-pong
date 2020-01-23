package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ezhdanovskiy/go-ping-pong/pkg/pinger"
	"github.com/ezhdanovskiy/go-ping-pong/pkg/rabbit"
	"github.com/pkg/errors"
)

func main() {
	var n = flag.Int("n", 0, "count of pings")
	var transport = flag.String("t", "rabbit", "transport for pings")
	flag.Parse()

	var broker pinger.Broker
	switch *transport {
	case "rabbit":
		//url := fmt.Sprintf("amqp://%s:%s@%s", os.Getenv("AMQP_URL"), os.Getenv("AMQP_USER"), os.Getenv("AMQP_PASS"))
		client, err := rabbit.NewClient("amqp://guest:guest@localhost:5672")
		if err != nil {
			log.Fatalf("Failed to create AMQP client: %s", err)
		}
		defer client.Close()
		broker = client
	default:
		fmt.Printf("Unknown transport: %s\n", *transport)
		return
	}

	srv := pinger.Service{
		Broker: broker,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("interrupt signal")
		cancel()
	}()

	stat, err := srv.Run(ctx, *n)
	if err != nil && errors.Cause(err) != context.Canceled {
		log.Print("pinger return error:", err)
	}
	log.Printf("stat: %+v", stat)
}
