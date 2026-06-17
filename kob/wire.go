package kob

import (
	"context"
	"log/slog"
	"sync"
)

type Wire struct {
	State          chan bool
	stationsMu     sync.Mutex
	stationCancels map[string]context.CancelFunc
}

func NewWire(ctx context.Context) *Wire {
	wire := &Wire{
		State:          make(chan bool, 1),
		stationCancels: make(map[string]context.CancelFunc),
	}
	wire.State <- false
	return wire
}

func (w *Wire) Close() {
	slog.Debug("closing wire")
	w.stationsMu.Lock()
	defer w.stationsMu.Unlock()
	for stationID, cancel := range w.stationCancels {
		cancel()
		delete(w.stationCancels, stationID)
	}
	close(w.State)
}

func (w *Wire) Connect(station *Station) {
	w.stationsMu.Lock()
	defer w.stationsMu.Unlock()
	_, exists := w.stationCancels[station.ID()]
	if exists {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w.stationCancels[station.ID()] = cancel
	go w.handleStation(ctx, station)
}

func (w *Wire) Disconnect(station *Station) {
	w.stationsMu.Lock()
	defer w.stationsMu.Unlock()
	cancel, exists := w.stationCancels[station.ID()]
	if exists {
		cancel()
		delete(w.stationCancels, station.ID())
	}
}

func (w *Wire) handleStation(ctx context.Context, station *Station) {
	for ctx.Err() == nil {
		select {
		case <-ctx.Done():
			return
		case state := <-station.State:
			slog.Debug("station state changed", "stationID", station.ID(), "state", state)
			if ctx.Err() != nil {
				slog.Debug("context cancelled, stopping station handler", "stationID", station.ID())
				return
			}
			w.State <- state
		}
	}
	slog.Debug("station handler stopped", "stationID", station.ID())
}
