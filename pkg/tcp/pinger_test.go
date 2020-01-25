package tcp

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pingAddr = ":3333"
	pongAddr = ":3334"
	pingID   = 1
)

func TestPinger_Pongs(t *testing.T) {
	runPongerMock(t)

	ctx, cancel := context.WithCancel(context.Background())

	p, err := NewPinger(ctx, pingAddr, pongAddr)
	require.NoError(t, err)
	defer p.Close()

	pongs, err := p.Pongs()
	require.NoError(t, err)

	err = p.Ping(pingID)
	require.NoError(t, err)

	id := <-pongs
	assert.Equal(t, pingID, id)

	err = p.Ping(pingID + 1)
	require.NoError(t, err)

	id = <-pongs
	assert.Equal(t, pingID+1, id)

	cancel()
}

func runPongerMock(t *testing.T) {
	listener, err := net.Listen("tcp", pingAddr)
	require.NoError(t, err)

	go func() {
		defer listener.Close()

		conn, err := listener.Accept()
		require.NoError(t, err)
		defer conn.Close()

		var pongConn net.Conn
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			require.NoError(t, err)

			if pongConn == nil {
				pongConn, err = net.Dial("tcp", pongAddr)
				require.NoError(t, err)
				defer pongConn.Close()
			}

			_, err = pongConn.Write(buf[:n])
			require.NoError(t, err)
		}
	}()
}
