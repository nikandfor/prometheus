package prometheus

import (
	"fmt"
	"sync"
	"time"

	"nikand.dev/go/quantile"
)

type (
	SummaryOpts = Opts

	SummaryVec = Vector[*Summary]

	Summary struct {
		d Desc

		nano func() int64

		mu sync.Mutex

		sum   float64
		count float64

		ss quantile.TDMulti

		granula int64 // const
		last    int

		qs    []float64
		qres  []float64
		qtext []string
		lsbuf Labels
	}

	summaryAllocator struct{}
)

func (a summaryAllocator) New() *Summary {
	s := &Summary{}
	s.initCounter()

	return s
}

func NewSummaryVec(o SummaryOpts, labelNames []string) *SummaryVec {
	a := summaryAllocator{}

	return NewVector(o, SummaryType, labelNames, a)
}

func NewSummary(o SummaryOpts) *Summary {
	s := &Summary{
		d: NewDescFromOpts(o, SummaryType),
	}

	s.initCounter()

	return s
}

func (m *Summary) initCounter() {
	const Granila = 5 * time.Second
	const Window = 15 * time.Second
	const N = Window / Granila

	m.qs = []float64{0.9, 0.99, 1}
	m.qres = make([]float64, len(m.qs))
	m.qtext = make([]string, len(m.qs))

	for i, q := range m.qs {
		m.qtext[i] = fmt.Sprintf("%v", q)
	}

	m.ss = make(quantile.TDMulti, N)

	for i := range m.ss {
		m.ss[i] = quantile.NewTDHighBiased(0.01, 1024)
	}

	m.granula = int64(Granila)
}

func (m *Summary) Collect(w Writer) error {
	err := m.d.CollectHeader(w)
	if err != nil {
		return err
	}

	return m.CollectMetric(w, nil)
}

func (m *Summary) CollectMetric(w Writer, ls Labels) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	last := len(ls)

	if last >= cap(m.lsbuf) {
		m.lsbuf = make(Labels, last+1)
	}

	copy(m.lsbuf, ls)
	m.lsbuf = m.lsbuf[:last+1]
	m.lsbuf[last].Name = "quantile"

	m.ss.QueryMulti(m.qs, m.qres)

	for i, v := range m.qres {
		m.lsbuf[last].Value = m.qtext[i]

		err := w.Write(v, "", m.lsbuf)
		if err != nil {
			return err
		}
	}

	m.lsbuf = m.lsbuf[:last]

	err := w.Write(m.sum, "sum", m.lsbuf)
	if err != nil {
		return err
	}

	err = w.Write(m.count, "count", m.lsbuf)
	if err != nil {
		return err
	}

	return nil
}

func (m *Summary) Observe(v float64) {
	if m == nil {
		return
	}

	g := m.curgranula()

	defer m.mu.Unlock()
	m.mu.Lock()

	if m.last == 0 {
		m.last = g
	}

	for i := m.last + 1; i <= g; i++ {
		m.ss[m.si(i)].Reset()
	}

	m.last = g

	m.ss[m.si(g)].Insert(v)

	m.sum += v
	m.count++
}

func (m *Summary) curgranula() int {
	var ts int64

	if m.nano != nil {
		ts = m.nano()
	} else {
		ts = time.Now().UnixNano()
	}

	return int(ts / m.granula)
}
func (m *Summary) si(i int) int { return i % len(m.ss) }

func csel[T any](c bool, x, y T) T {
	if c {
		return x
	}

	return y
}
