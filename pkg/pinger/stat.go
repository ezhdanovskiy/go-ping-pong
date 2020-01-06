package pinger

import (
	"log"
	"math"
	"sync"
	"time"
)

type PingSessionStat struct {
	Sent     int
	Received int
	MinTime  time.Duration
	AvgTime  time.Duration
	MaxTime  time.Duration
}

func newStat() stat {
	return stat{
		PingSessionStat: PingSessionStat{
			MinTime: time.Duration(math.MaxInt64),
		},
		pings: make(map[int]time.Time),
	}
}

type stat struct {
	PingSessionStat

	mtx    sync.Mutex
	pings  map[int]time.Time
	durSum time.Duration
}

func (s *stat) Send(id int) {
	log.Printf("Send(%v)", id)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.Sent++
	s.pings[id] = time.Now()
}

func (s *stat) Receive(id int) {
	log.Printf("Receive(%v)", id)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.Received++
	if sent, ok := s.pings[id]; ok {
		dur := time.Now().Sub(sent)
		if s.MinTime > dur {
			s.MinTime = dur
		}
		if s.MaxTime < dur {
			s.MaxTime = dur
		}
		s.durSum += dur
	}
}

func (s *stat) Stat() *PingSessionStat {
	log.Print("Stat()")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.Received > 0 {
		s.AvgTime = s.durSum / time.Duration(s.Received)
	}
	return &s.PingSessionStat
}
