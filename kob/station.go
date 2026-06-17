package kob

import (
	"log/slog"
	"math"
	"sync"
	"time"
)

type Station struct {
	id              string
	version         string
	pendingEventsMu sync.Mutex
	pendingEvents   []int32
	State           chan bool
	closed          bool
}

func NewStation(id, version string) *Station {
	station := &Station{
		id:            id,
		version:       version,
		State:         make(chan bool, 1),
		pendingEvents: make([]int32, 0),
	}
	go station.processLoop()
	return station
}

func (s *Station) Close() {
	slog.Debug("closing station", "stationID", s.id)
	s.closed = true
}

func (s *Station) processLoop() {
	for !s.closed {
		s.pendingEventsMu.Lock()
		if len(s.pendingEvents) > 0 {
			pending := s.pendingEvents[0]
			if pending < -3000 && len(s.pendingEvents) > 1 {
				if s.pendingEvents[1] == 2 {
					// special case for break (consume two events)
					s.pendingEvents = s.pendingEvents[2:]
					s.State <- false
					continue
				}
			} else {
				s.pendingEvents = s.pendingEvents[1:]
				switch pending {
				default:
					s.State <- pending < 0
				}
			}
			s.pendingEventsMu.Unlock()
			time.Sleep(time.Duration(math.Abs(float64(pending))) * time.Millisecond)
		} else {
			s.pendingEventsMu.Unlock()
		}
	}
}

func (s *Station) ID() string {
	return s.id
}

func (s *Station) Version() string {
	return s.version
}

func (s *Station) PushCodeList(codeList []int32) {
	if len(codeList) == 0 {
		return
	}

	s.pendingEventsMu.Lock()
	s.pendingEvents = append(s.pendingEvents, codeList...)
	s.pendingEventsMu.Unlock()
}
