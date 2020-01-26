package tcp

import (
	"context"
	"net"
	"sync"
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
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	runPongerMock(ctx, &wg, t)

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
	wg.Wait()
}

func runPongerMock(ctx context.Context, wg *sync.WaitGroup, t *testing.T) {
	listenConfig := &net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", pingAddr)
	require.NoError(t, err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer listener.Close()

		conn, err := listener.Accept()
		require.NoError(t, err)

		var pongConn net.Conn
		buf := make([]byte, 1024)
		for i := 0; i < 2; i++ {
			n, err := conn.Read(buf)
			//log.Printf("conn.Read: %s", buf)
			require.NoError(t, err)

			if pongConn == nil {
				pongConn, err = net.Dial("tcp", pongAddr)
				require.NoError(t, err)
			}

			_, err = pongConn.Write(buf[:n])
			require.NoError(t, err)
		}
		pongConn.Close()
		conn.Close()
	}()
}
