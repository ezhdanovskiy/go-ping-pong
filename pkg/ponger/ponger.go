package ponger

import (
	"context"
	"log"

	"github.com/pkg/errors"
)

type Broker interface {
	Pings() (<-chan int, error)
	Pong(id int) error
}

type Service struct {
	Broker Broker
}

func (s *Service) Run(ctx context.Context) (n int, err error) {
	pings, err := s.Broker.Pings()
	if err != nil {
		return 0, errors.Wrap(err, "failed to gep pings chan")
	}

	for {
		select {
		case <-ctx.Done():
			log.Print(ctx.Err())
			return n, ctx.Err()
		case pingID := <-pings:
			err := s.Broker.Pong(pingID)
			if err != nil {
				log.Printf("Failed to send ping %v: %s", pingID, err)
				break
			}
			n++
		}
	}
}
