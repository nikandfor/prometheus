package promhttp

import (
	"net/http"
	"sync"

	"nikand.dev/go/prometheus"
)

type (
	server struct {
		mu sync.Mutex

		c  prometheus.Collector
		bw prometheus.BufWriter
	}
)

func Handler() http.Handler {
	return &server{
		c: prometheus.DefaultRegistry,
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer s.mu.Unlock()
	s.mu.Lock()

	s.bw.Reset()

	err := s.c.Collect(&s.bw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	_, _ = s.bw.WriteTo(w)
}
