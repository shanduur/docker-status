# Container Status Dashboard

A simple web and Prometheus endpoint to monitor Docker container status.

## Overview

* Exposes container metrics at `/metrics` in Prometheus format.
* Serves a live-updating web dashboard at `/` via WebSocket (`/ws`).
* Uses Tailwind CSS for basic styling.

## Metrics

### `docker_container_status_info`

Info with the following labels:

* `id` – Container ID
* `name` – Container name
* `image` – Container image
* `status` – Status string (e.g. "Up 10 minutes")
* `state` – Container state (e.g. "running")
