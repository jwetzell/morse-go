package kob

import (
	"context"
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
	ctx             context.Context
	latched         bool
}

func NewStation(ctx context.Context, id, version string) *Station {
	station := &Station{
		ctx:           ctx,
		id:            id,
		version:       version,
		State:         make(chan bool, 1),
		pendingEvents: make([]int32, 0),
	}
	go station.processLoop()
	return station
}

func (s *Station) sendState(state bool) {
	select {
	case s.State <- state:
	default:
	}
}

func (s *Station) processLoop() {
	for s.ctx.Err() == nil {
		s.pendingEventsMu.Lock()
		if len(s.pendingEvents) > 0 {
			pending := s.pendingEvents[0]
			if pending < 0 && len(s.pendingEvents) > 1 {
				switch s.pendingEvents[1] {
				case 2:
					s.latched = false
					s.pendingEvents = s.pendingEvents[2:]
					s.sendState(false)
					s.pendingEventsMu.Unlock()
					continue
				}
			}
			s.pendingEvents = s.pendingEvents[1:]
			s.pendingEventsMu.Unlock()
			time.Sleep(time.Duration(math.Abs(float64(pending))) * time.Millisecond)
			switch pending {
			case 1:
				s.latched = true
				s.sendState(true)
			case 2:
				s.latched = false
				s.sendState(false)
			default:
				if !s.latched {
					s.sendState(pending > 0)
				}
			}
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
