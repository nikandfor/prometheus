package prometheus

import (
	"strconv"
	"sync"
	"time"

	"nikand.dev/go/quantile"
)

type (
	SummaryOpts struct {
		Quantiles []float64
	}

	Summary struct {
		d Desc

		mu sync.Mutex

		qs, res []float64 // quantile calc tmp vars
		qtext   []string  // encoded to text quantile values

		ln, lv []string // tmp vars to add quantile="0.9" to metric labels

		sum   float64
		count int

		// end of mu

		Now func() time.Time // mockable time source

		// counter

		// protected by par.mu

		par *Summary

		ss []*quantile.TDigest

		granula int64
		last    int

		// end of par.mu
	}

	SummaryVec = Vector[*Summary]

	summaryAlloc struct {
		par *Summary
	}
)

func NewSummary(opts Opts, sumo SummaryOpts) *Summary {
	m := &Summary{
		d: opts.desc(SummaryType),
	}

	m.initMetric(sumo.Quantiles)
	m.initCounter()

	return m
}

func NewSummaryVec(opts Opts, sumo SummaryOpts, labelNames []string) *SummaryVec {
	m := &Summary{}
	m.initMetric(sumo.Quantiles)

	return newVector[*Summary](opts, SummaryType, labelNames, summaryAlloc{m})
}

func (v *Summary) initMetric(qs []float64) {
	v.qs = append(v.qs[:0], qs...)
	v.res = append(v.res[:0], qs...)
	v.qtext = make([]string, len(qs))

	for i, q := range qs {
		v.qtext[i] = strconv.FormatFloat(q, 'g', 5, 64)
	}
}

func (s *Summary) initCounter() {
	const Granula = 3 * time.Second
	const Window = 15 * time.Second

	s.ss = make([]*quantile.TDigest, Window/Granula)

	for i := range s.ss {
		s.ss[i] = quantile.NewExtremesBiased(0.01, 4096)
	}

	s.granula = int64(Granula)
}

func (v *Summary) Observe(x float64) {
	if v == nil {
		return
	}

	p := first(v.par, v)

	now := p.Now
	if now == nil {
		now = time.Now
	}

	ts := now().UnixNano()
	g := int(ts / v.granula)

	defer p.mu.Unlock()
	p.mu.Lock()

	if v.last == 0 {
		v.last = g
	}

	for i := v.last + 1; i <= g; i++ {
		v.ss[v.si(i)].Reset()
	}

	v.last = g

	v.ss[v.si(g)].Insert(x)

	p.sum += x
	p.count++
}

func (v *Summary) si(i int) int { return i % len(v.ss) }

func (v *Summary) Collect(w Writer) error {
	err := v.d.WriteHeader(w)
	if err != nil {
		return err
	}

	return v.writeMetric(w, nil, nil)
}

func (v *Summary) writeMetric(w Writer, ln, lv []string) error {
	p := first(v.par, v)

	defer p.mu.Unlock()
	p.mu.Lock()

	p.ln = append(p.ln[:0], ln...)
	p.lv = append(p.lv[:0], lv...)

	last := len(p.ln)

	p.ln = append(p.ln, "quantile")
	p.lv = append(p.lv, "")

	quantile.QueryMulti(p.qs, p.res, v.ss...)

	for i := range p.qs {
		p.lv[last] = p.qtext[i]

		err := w.Metric("", p.res[i], p.ln, p.lv)
		if err != nil {
			return err
		}
	}

	err := w.Metric("sum", p.sum, p.ln[:last], p.lv[:last])
	if err != nil {
		return err
	}

	err = w.Metric("count", float64(p.count), p.ln[:last], p.lv[:last])
	if err != nil {
		return err
	}

	return nil
}

func (a summaryAlloc) new() *Summary {
	s := &Summary{par: a.par}
	s.initCounter()
	return s
}

func first[T comparable](x ...T) T {
	var zero T

	for _, x := range x {
		if x != zero {
			return x
		}
	}

	return zero
}
