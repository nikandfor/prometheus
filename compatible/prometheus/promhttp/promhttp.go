package promhttp

import (
	"net/http"

	"nikand.dev/go/prometheus"
)

func Handler() http.Handler {
	return prometheus.DefaultRegistry
}
