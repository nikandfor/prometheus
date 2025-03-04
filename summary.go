package prometheus

import (
	"strconv"
	"sync"

	"nikand.dev/go/quantile"
)

type (
	SummaryOpts struct {
		Opts

		Quantiles []float64
	}

	SummaryMetric struct {
		d Desc

		mu      sync.Mutex
		qs, res []float64
		qtext   []string
		ln, lv  []string

		sum   float64
		count int

		Summary
	}

	Summary struct {
		par *SummaryMetric

		s *quantile.TDigest
	}

	SummaryVec = Vector[*Summary, summaryAlloc]

	Quantile interface {
		Insert(v float64)
		Query(q float64) float64
	}

	QuantileMulti interface {
		QueryMulti(qs, res []float64)
	}

	summaryAlloc struct {
		par *SummaryMetric
	}
)

func NewSummary(opts SummaryOpts) *SummaryMetric {
	m := &SummaryMetric{
		d: opts.desc(SummaryType),
	}

	m.Summary = *summaryAlloc{m}.new()
	m.init(opts.Quantiles)

	return m
}

func NewSummaryVec(opts SummaryOpts, labelNames []string) *SummaryVec {
	m := &SummaryMetric{}
	m.init(opts.Quantiles)

	return newVector[*Summary, summaryAlloc](opts.Opts, SummaryType, labelNames, summaryAlloc{m})
}

func (v *SummaryMetric) init(qs []float64) {
	v.qs = append(v.qs[:0], qs...)
	v.res = append(v.res[:0], qs...)
	v.qtext = make([]string, len(qs))

	for i, q := range qs {
		v.qtext[i] = strconv.FormatFloat(q, 'g', 5, 64)
	}
}

func (v *Summary) Observe(x float64) {
	defer v.par.mu.Unlock()
	v.par.mu.Lock()

	v.s.Insert(x)

	v.par.sum += x
	v.par.count++
}

func (v *SummaryMetric) Collect(w Writer) error {
	err := v.d.WriteHeader(w)
	if err != nil {
		return err
	}

	return v.writeMetric(w, nil, nil)
}

func (v *Summary) writeMetric(w Writer, ln, lv []string) error {
	defer v.par.mu.Unlock()
	v.par.mu.Lock()

	v.par.ln = append(v.par.ln[:0], ln...)
	v.par.lv = append(v.par.lv[:0], lv...)

	last := len(v.par.ln)

	v.par.ln = append(v.par.ln, "quantile")
	v.par.lv = append(v.par.lv, "")

	v.s.QueryMulti(v.par.qs, v.par.res)

	for i := range v.par.qs {
		v.par.lv[last] = v.par.qtext[i]

		err := w.Metric("", v.par.res[i], v.par.ln, v.par.lv)
		if err != nil {
			return err
		}
	}

	err := w.Metric("sum", v.par.sum, v.par.ln[:last], v.par.lv[:last])
	if err != nil {
		return err
	}

	err = w.Metric("count", float64(v.par.count), v.par.ln[:last], v.par.lv[:last])
	if err != nil {
		return err
	}

	return nil
}

func (a summaryAlloc) new() *Summary {
	const N = 512

	return &Summary{
		par: a.par,

		s: quantile.NewTDigest(N, quantile.TDigestEpsilon(N)),
	}
}
