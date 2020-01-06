package pinger

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinger_OnePing(t *testing.T) {
	p := Pinger{
		PingSession: &pingSessionMock{ch: make(chan int)},
		Wait:        100 * time.Millisecond,
	}

	stat, err := p.Run(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, 1, stat.Sent)
	assert.Equal(t, 1, stat.Received)
	fmt.Printf("stat: %+v\n", stat)
}

func TestPinger_ThreePings(t *testing.T) {
	p := Pinger{
		PingSession: &pingSessionMock{ch: make(chan int)},
		Wait:        100 * time.Millisecond,
	}

	stat, err := p.Run(context.Background(), 3)
	require.NoError(t, err)
	assert.Equal(t, 3, stat.Sent)
	assert.Equal(t, 3, stat.Received)
	fmt.Printf("stat: %+v\n", stat)
}

func TestPinger_Cancel(t *testing.T) {
	p := Pinger{
		PingSession: &pingSessionMock{ch: make(chan int, 1)},
		Wait:        200 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()

	stat, err := p.Run(ctx, 3)
	require.Equal(t, "context canceled", errors.Cause(err).Error())
	assert.Equal(t, 1, stat.Sent)
	assert.Equal(t, 1, stat.Received)
	fmt.Printf("stat: %+v\n", stat)
}

type pingSessionMock struct {
	ch chan int
}

func (s *pingSessionMock) Ping(id int) error {
	rand.Seed(time.Now().UnixNano())
	dur := time.Duration(rand.Intn(100)) * time.Millisecond
	fmt.Printf("dur: %v\n", dur)

	time.Sleep(dur)
	s.ch <- id
	return nil
}

func (s *pingSessionMock) Pongs() (<-chan int, error) {
	return s.ch, nil
}
