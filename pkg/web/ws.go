package web

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shanduur/docker-status/pkg/store"
)

var upgrader = websocket.Upgrader{}

type StatsHandler struct {
	store *store.Store
}

func NewStatsHandler(store *store.Store) *StatsHandler {
	return &StatsHandler{
		store: store,
	}
}

func (sh *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	n := 1

	interval := r.URL.Query().Get("n")
	if interval != "" {
		np, err := strconv.Atoi(interval)
		if err == nil {
			n = np
			log.Printf("Ticking interval is set to %ds", n)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				log.Printf("WebSocket disconnected: %v", err)
				cancel()
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Second * time.Duration(n))
	defer ticker.Stop()

	stats := sh.store.Get()
	if err := conn.WriteJSON(stats); err != nil {
		log.Println("WebSocket write error:", err)
	}

	for {
		select {
		case <-ticker.C:
			stats := sh.store.Get()

			if err := conn.WriteJSON(stats); err != nil {
				log.Println("WebSocket write error:", err)
			}

		case <-ctx.Done():
			return
		}
	}
}
