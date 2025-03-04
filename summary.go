package prometheus

import (
	"strconv"
	"sync"
	"time"

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

		ss []*quantile.TDigest[quantile.ExtremesBias]

		granula int64
		last    int
	}

	SummaryVec = Vector[*Summary, summaryAlloc]

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
	ts := time.Now().UnixNano()
	g := int(ts / v.granula)

	defer v.par.mu.Unlock()
	v.par.mu.Lock()

	if v.last == 0 {
		v.last = g
	}

	for i := v.last + 1; i <= g; i++ {
		v.ss[v.si(i)].Reset()
	}

	v.last = g

	v.ss[v.si(g)].Insert(x)

	v.par.sum += x
	v.par.count++
}

func (v *Summary) si(i int) int { return i % len(v.ss) }

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

	quantile.QueryMulti(v.par.qs, v.par.res, v.ss...)

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
	const Granula = 3 * time.Second
	const Window = 15 * time.Second

	ss := make([]*quantile.TDigest[quantile.ExtremesBias], Window/Granula)

	for i := range ss {
		ss[i] = quantile.NewExtremesBiased(0.01, 4096)
	}

	return &Summary{
		par: a.par,

		ss: ss,

		granula: int64(Granula),
	}
}
