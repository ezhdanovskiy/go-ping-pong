package pinger

import (
	"context"
	"log"
	"math"
	"time"

	"github.com/pkg/errors"
)

type PingSession interface {
	Ping(id int) error
	Pongs() (<-chan int, error)
}

type Pinger struct {
	PingSession PingSession
	Wait        time.Duration
}

func (p *Pinger) Run(ctx context.Context, n int) (*PingSessionStat, error) {
	if n < 0 {
		n = math.MaxInt32
	}

	wait := p.Wait
	if wait == 0 {
		wait = time.Second
	}

	stat := newStat(wait)

	pongDone := make(chan error, 1)
	go func() {
		var err error
		defer func() { pongDone <- errors.Wrap(err, "pong receiver shutdown") }()

		pongs, err := p.PingSession.Pongs()
		if err != nil {
			log.Print("failed to get pongs chan")
			return
		}

		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case pong := <-pongs:
				stat.Receive(pong)
			}
		}
	}()

	for i := 0; i < n; i++ {
		err := p.PingSession.Ping(i)
		if err != nil {
			return nil, errors.Wrap(err, "failed to send ping")
		}
		stat.Send(i)

		select {
		case <-ctx.Done():
			return stat.Stat(), errors.Wrap(ctx.Err(), "ping shutdown")
		case err = <-pongDone:
			return stat.Stat(), err
		case <-time.After(wait):
			continue
		}
	}

	return stat.Stat(), nil
}
