package store

import (
	"slices"
	"sync"

	"github.com/shanduur/docker-status/pkg/model"
)

type Store struct {
	data map[string]model.ContainerStat
	mux  sync.RWMutex
}

func (s *Store) Save(data []model.ContainerStat) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.data == nil {
		s.data = make(map[string]model.ContainerStat)
	}

	clear(s.data)

	for _, stat := range data {
		s.data[stat.ID] = stat
	}
}

func (s *Store) Get() []model.ContainerStat {
	s.mux.RLock()
	defer s.mux.RUnlock()

	stats := make([]model.ContainerStat, 0, len(s.data))
	for _, stat := range s.data {
		stats = append(stats, stat)
	}

	slices.SortFunc(stats, func(a, b model.ContainerStat) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})

	return stats
}
