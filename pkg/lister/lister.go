package lister

import (
	"context"
	"log"
	"maps"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shanduur/monorepo/docker-status/pkg/model"
	"github.com/shanduur/monorepo/docker-status/pkg/store"
)

var (
	containerStatusInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "docker_container_status_info",
			Help: "Information about the status of Docker containers",
		},
		[]string{"id", "name", "image", "status", "state"},
	)
)

type Lister struct {
	cli   *client.Client
	store *store.Store
}

func New(store *store.Store) (*Lister, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}

	// Register the metric
	prometheus.MustRegister(containerStatusInfo)

	return &Lister{
		cli:   cli,
		store: store,
	}, nil
}

func (l *Lister) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	visitedContainers := make(map[string]model.ContainerStat)
	iterationVisited := make(map[string]model.ContainerStat)

	for {
		select {
		case <-ticker.C:
			clear(iterationVisited)

			containers, err := l.cli.ContainerList(ctx, container.ListOptions{All: true})
			if err != nil {
				log.Println("List error:", err)
				return
			}

			var data []model.ContainerStat
			for _, c := range containers {
				// update metric for the container
				containerStatusInfo.WithLabelValues(
					c.ID[:12],
					c.Names[0],
					c.Image,
					c.Status,
					c.State,
				).Set(1)

				ms := model.ContainerStat{
					ID:     c.ID[:12],
					Name:   c.Names[0],
					Image:  c.Image,
					Status: c.Status,
					State:  c.State,
				}

				data = append(data, ms)

				// mark the container as visited in this iteration
				iterationVisited[c.ID[:12]] = ms
			}

			l.store.Save(data)

			// remove containers that were not visited in this iteration
			for id := range visitedContainers {
				if data, found := iterationVisited[id]; !found {
					containerStatusInfo.DeleteLabelValues(
						data.ID,
						data.Name,
						data.Image,
						data.Status,
						data.State,
					)
				} else {
					// if any field has changed, drop the old metric
					if visitedContainers[id] != data {
						containerStatusInfo.DeleteLabelValues(
							visitedContainers[id].ID,
							visitedContainers[id].Name,
							visitedContainers[id].Image,
							visitedContainers[id].Status,
							visitedContainers[id].State,
						)
					}
				}
			}

			// clear the visited containers for the next iteration
			clear(visitedContainers)
			// copy the current iteration's visited containers to the main map
			maps.Copy(visitedContainers, iterationVisited)

		case <-ctx.Done():
			return
		}
	}
}
