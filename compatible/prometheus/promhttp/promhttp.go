package promhttp

import (
	"net/http"

	"nikand.dev/go/prometheus/promhttp"
)

func Handler() http.Handler {
	return promhttp.Handler()
}
