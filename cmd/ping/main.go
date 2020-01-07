package main

import (
	"context"
	"flag"
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
	flag.Parse()

	url := "amqp://guest:guest@localhost:5672"
	//url := fmt.Sprintf("amqp://%s:%s@%s", os.Getenv("AMQP_URL"), os.Getenv("AMQP_USER"), os.Getenv("AMQP_PASS"))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("interrupt signal")
		cancel()
	}()

	client, err := rabbit.NewClient(url)
	if err != nil {
		log.Fatalf("Failed to create AMQP client: %s", err)
	}
	defer client.Close()

	srv := pinger.Service{
		Broker: client,
	}

	stat, err := srv.Run(ctx, *n)
	if err != nil && errors.Cause(err) != context.Canceled {
		log.Print("pinger return error:", err)
	}
	log.Printf("stat: %+v", stat)
}
