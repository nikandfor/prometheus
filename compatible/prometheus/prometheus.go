package prometheus

import (
	pp "nikand.dev/go/prometheus"
)

type (
	Desc   pp.Desc
	Metric interface{}
	Labels map[string]string

	CounterOpts = pp.CounterOpts
	Counter     = *pp.Counter
	CounterVec  = pp.CounterVec

	Summary    = *pp.Summary
	SummaryVec = pp.SummaryVec

	SummaryOpts struct {
		Namespace string
		Subsystem string
		Name      string

		Help string

		//	ConstLabels Labels

		Objectives map[float64]float64
	}
)

func NewCounter(opts CounterOpts) Counter {
	return pp.NewCounter(opts)
}

func NewCounterVec(opts CounterOpts, labels []string) *CounterVec {
	return pp.NewCounterVec(opts, labels)
}

func NewSummary(opts SummaryOpts) Summary {
	return pp.NewSummary(summaryOpts(opts))
}

func NewSummaryVec(opts SummaryOpts, labels []string) *SummaryVec {
	o := summaryOpts(opts)

	return pp.NewSummaryVec(o, labels)
}

func summaryOpts(opts SummaryOpts) pp.SummaryOpts {
	return pp.SummaryOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		//	ConstLabels: convertLabels(opts.ConstLabels),
	}
}

func convertLabels(ls Labels) []pp.Label {
	r := make([]pp.Label, len(ls))
	i := 0

	for n, v := range ls {
		r[i] = pp.Label{
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
