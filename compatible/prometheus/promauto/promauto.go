package promauto

import (
	"github.com/prometheus/client_golang/prometheus"
)

func NewCounter(opts prometheus.CounterOpts) prometheus.Counter {
	c := prometheus.NewCounter(opts)

	return c
}

func NewCounterVec(opts prometheus.CounterOpts, labels []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(opts, labels)

	return c
}

func NewSummary(opts prometheus.SummaryOpts) prometheus.Summary {
	c := prometheus.NewSummary(opts)

	return c
}

func NewSummaryVec(opts prometheus.SummaryOpts, labels []string) *prometheus.SummaryVec {
	c := prometheus.NewSummaryVec(opts, labels)

	return c
}
