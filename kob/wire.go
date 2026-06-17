package kob

import (
	"math"
	"sync"
	"time"
)

type Wire struct {
	State             chan bool
	pendingMessagesMu sync.Mutex
	pendingMesssages  []int32
}

func NewWire() *Wire {
	wire := &Wire{
		State: make(chan bool, 1),
	}
	go wire.processLoop()
	return wire
}

func (w *Wire) processLoop() {

	for {
		w.pendingMessagesMu.Lock()
		if len(w.pendingMesssages) > 0 {
			pending := w.pendingMesssages[0]
			w.pendingMesssages = w.pendingMesssages[1:]
			w.State <- pending > 0
			w.pendingMessagesMu.Unlock()
			time.Sleep(time.Duration(math.Abs(float64(pending))) * time.Millisecond)
		} else {
			w.pendingMessagesMu.Unlock()
		}
	}
}
func (w *Wire) RegisterCodeList(codeList []int32) {
	if len(codeList) == 0 {
		return
	}

	w.pendingMessagesMu.Lock()
	w.pendingMesssages = append(w.pendingMesssages, codeList...)
	w.pendingMessagesMu.Unlock()
}
