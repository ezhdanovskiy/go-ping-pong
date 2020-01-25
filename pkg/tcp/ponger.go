package tcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"

	"github.com/ezhdanovskiy/go-ping-pong/pkg/models"
	"github.com/pkg/errors"
)

func NewPonger(ctx context.Context, pingAddr, pongAddr string) (*ponger, error) {
	listenConfig := &net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", pingAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start listening")
	}
	return &ponger{
		ctx:          ctx,
		pingListener: listener,
		pongAddr:     pongAddr,
	}, nil
}

type ponger struct {
	ctx          context.Context
	pingListener net.Listener
	pongAddr     string
	pongConn     net.Conn
}

func (p *ponger) Pings() (<-chan int, error) {
	chIDs := make(chan int)

	go func() {
		for {
			conn, err := p.pingListener.Accept()
			if err != nil {
				break
			}

			go func(pingConn net.Conn) {
				// Make a buffer to hold incoming data.
				buf := make([]byte, 1024)

				for {
					// Read the incoming connection into the buffer.
					n, err := pingConn.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Printf("Error reading pings: %s", err)
						}
						break
					}
					log.Printf("received: %s", buf[:n])

					var event models.Event
					err = json.Unmarshal(buf[:n], &event)
					if err != nil {
						log.Printf("Failed to unmarshal ping event: %s", buf[:n])
					}
					chIDs <- event.ID
				}

				err = pingConn.Close()
				if err != nil {
					log.Printf("Failed to cloce ping connection: %s", err)
				}

				err = p.pongConn.Close()
				if err != nil {
					log.Printf("Failed to cloce pong connection: %s", err)
				}
				p.pongConn = nil
			}(conn)
		}
	}()

	return chIDs, nil
}

func (p *ponger) Pong(id int) error {
	if p.pongConn == nil {
		var err error
		p.pongConn, err = net.Dial("tcp", p.pongAddr)
		if err != nil {
			return errors.Wrap(err, "failed to dial")
		}
	}

	bytes, err := json.Marshal(&models.Event{ID: id})
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	_, err = p.pongConn.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "failed to write to socket")
	}

	return nil
}

func (p *ponger) Close() error {
	return p.pingListener.Close()
}
