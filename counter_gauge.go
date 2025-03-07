package prometheus

import "sync"

type (
	CounterOpts = Opts
	GaugeOpts   = Opts

	CounterVec = Vector[*Counter]
	GaugeVec   = Vector[*Gauge]

	Counter struct {
		d Desc

		mu sync.Mutex

		v float64
	}

	Gauge Counter
)

func NewCounterVec(o CounterOpts, labelNames []string) *CounterVec {
	var a BasicAllocator[Counter]

	return NewVector(o, CounterType, labelNames, a)
}

func NewCounter(o CounterOpts) *Counter {
	return &Counter{
		d: NewDescFromOpts(o, CounterType),
	}
}

func (m *Counter) Inc() { m.Add(1) }

func (m *Counter) Add(v float64) {
	if m == nil {
		return
	}

	defer m.mu.Unlock()
	m.mu.Lock()

	m.v += v
}

func (m *Counter) Collect(w Writer) error {
	err := m.d.CollectHeader(w)
	if err != nil {
		return err
	}

	return m.CollectMetric(w, nil)
}

func (m *Counter) CollectMetric(w Writer, ls Labels) error {
	defer m.mu.Unlock()
	m.mu.Lock()

	err := w.Write(m.v, "", ls)
	if err != nil {
		return err
	}

	return nil
}

func NewGaugeVec(o GaugeOpts, labelNames []string) *GaugeVec {
	var a BasicAllocator[Gauge]

	return NewVector(o, GaugeType, labelNames, a)
}

func NewGauge(o GaugeOpts) *Gauge {
	return &Gauge{
		d: NewDescFromOpts(o, GaugeType),
	}
}

func (m *Gauge) Inc()          { m.Add(1) }
func (m *Gauge) Dec()          { m.Add(-1) }
func (m *Gauge) Sub(v float64) { m.Add(-v) }

func (m *Gauge) Add(v float64) {
	if m == nil {
		return
	}

	defer m.mu.Unlock()
	m.mu.Lock()

	m.v += v
}

func (m *Gauge) Set(v float64) {
	if m == nil {
		return
	}

	defer m.mu.Unlock()
	m.mu.Lock()

	m.v = v
}

func (m *Gauge) Collect(w Writer) error {
	return (*Counter)(m).Collect(w)
}

func (m *Gauge) CollectMetric(w Writer, ls Labels) error {
	return (*Counter)(m).CollectMetric(w, ls)
}
