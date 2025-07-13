package web

import (
	"log"
	"net/http"
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
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	ctx := r.Context()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := sh.store.Get()

			if err := conn.WriteJSON(stats); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
