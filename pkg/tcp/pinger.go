package tcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"

	"github.com/ezhdanovskiy/go-ping-pong/pkg/models"
	"github.com/pkg/errors"
)

func NewPinger(ctx context.Context, pingAddr, pongAddr string) (*pinger, error) {
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", pingAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial")
	}
	return &pinger{
		conn:     conn,
		pongAddr: pongAddr,
	}, nil
}

type pinger struct {
	conn     net.Conn
	pongAddr string
}

func (p *pinger) Ping(id int) error {
	if p.conn == nil {
		return errors.New("connection not established")
	}

	bytes, err := json.Marshal(&models.Event{ID: id})
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	_, err = p.conn.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "failed to write to socket")
	}

	return nil
}

func (p *pinger) Pongs() (<-chan int, error) {
	chIDs := make(chan int)

	listener, err := net.Listen("tcp", p.pongAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to start listening")
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error accepting pongs: ", err.Error())
				os.Exit(1)
			}

			go func(conn net.Conn) {
				defer conn.Close()

				// Make a buffer to hold incoming data.
				buf := make([]byte, 1024)

				for {
					// Read the incoming connection into the buffer.
					n, err := conn.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Printf("Error reading pongs: %s", err)
						}
						break
					}

					var event models.Event
					err = json.Unmarshal(buf[:n], &event)
					if err != nil {
						log.Printf("Failed to unmarshal pong event: %s", buf[:n])
					}
					chIDs <- event.ID
				}
			}(conn)

		}
	}()

	return chIDs, nil

}

func (p *pinger) Close() error {
	return p.conn.Close()
}
