package prometheus

import "nikand.dev/go/prometheus"

type (
	Desc   = prometheus.Desc
	Metric interface{}
	Labels map[string]string

	CounterOpts = prometheus.CounterOpts
	Counter     = *prometheus.Counter
	CounterVec  = prometheus.CounterVec

	Summary    = *prometheus.Summary
	SummaryVec = prometheus.SummaryVec

	SummaryOpts struct {
		Namespace string
		Subsystem string
		Name      string

		Help string

		Labels Labels

		Objectives map[float64]float64
	}
)

func NewCounter(opts CounterOpts) Counter {
	return prometheus.NewCounter(opts)
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	return prometheus.NewCounterVec(opts, labels)
}

func NewSummary(opts SummaryOpts) Summary {
	return prometheus.NewSummary(convertSummaryOpts(opts))
}

func NewSummaryVec(opts SummaryOpts, labels []string) *SummaryVec {
	o := convertSummaryOpts(opts)

	return prometheus.NewSummaryVec(o, labels)
}

func convertSummaryOpts(opts SummaryOpts) prometheus.SummaryOpts {
	return prometheus.Opts{
		Namespace:   opts.Namespace,
		Subsystem:   opts.Subsystem,
		Name:        opts.Name,
		ConstLabels: convertLabels(opts.Labels),
	}
}

func convertLabels(ls Labels) []prometheus.Label {
	r := make([]prometheus.Label, len(ls))
	i := 0

	for n, v := range ls {
		r[i] = prometheus.Label{
			Name:  n,
			Value: v,
		}

		i++
	}

	return r
}

func convertQuantiles(o map[float64]float64) []float64 {
	r := make([]float64, len(o))
	i := 0

	for q := range o {
		r[i] = q
		i++
	}

	return r
}
