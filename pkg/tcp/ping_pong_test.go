package tcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingPong(t *testing.T) {
	const (
		pingAddr = ":3343"
		pongAddr = ":3344"
	)
	ctx, cancel := context.WithCancel(context.Background())

	ponger, err := NewPonger(ctx, pingAddr, pongAddr)
	require.NoError(t, err)

	pings, err := ponger.Pings()
	require.NoError(t, err)

	pinger, err := NewPinger(ctx, pingAddr, pongAddr)
	require.NoError(t, err)

	err = pinger.Ping(pingID)
	require.NoError(t, err)

	id := <-pings
	assert.Equal(t, pingID, id)

	pongs, err := pinger.Pongs()
	require.NoError(t, err)

	err = ponger.Pong(id)
	require.NoError(t, err)

	id = <-pongs
	assert.Equal(t, pingID, id)

	cancel()

	err = ponger.Close()
	require.NoError(t, err)

	err = pinger.Close()
	require.NoError(t, err)
}
