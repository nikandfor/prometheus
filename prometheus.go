package prometheus

import "sync"

type (
	Type int

	// Opts is a prometheus Opts [1].
	//
	// [1] https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Opts
	Opts struct {
		Namespace string
		Subsystem string
		Name      string

		Help string

		Labels []Label
	}

	Label struct {
		Name  string
		Value string
	}

	GaugeOpts = Opts

	GaugeMetric struct {
		d Desc

		Gauge
	}

	Gauge struct {
		mu sync.Mutex

		v float64
	}

	CounterOpts = Opts

	CounterMetric struct {
		d Desc

		Counter
	}

	Counter struct {
		mu sync.Mutex

		v float64
	}

	GaugeVec   = Vector[*Gauge, gnew[Gauge]]
	CounterVec = Vector[*Counter, gnew[Counter]]

	Vector[T writerMetric, A alloc[T]] struct {
		d      Desc
		labels []string // names

		a A

		mu sync.Mutex

		list []vecElem[T]
		hash map[uintptr][]int
	}

	vecElem[T writerMetric] struct {
		labels []string // values

		v T
	}

	writerMetric interface {
		writeMetric(w Writer, ln, lv []string) error
	}

	alloc[T any] interface {
		new() T
	}
)

const (
	GaugeType Type = iota
	CounterType
	SummaryType
)

func (t Type) String() string { return typeNames[t] }

var typeNames = []string{"gauge", "counter", "summary"}

func NewGauge(opts GaugeOpts) *GaugeMetric {
	return &GaugeMetric{
		d: opts.desc(GaugeType),
	}
}

func NewGaugeVec(opts GaugeOpts, labelNames []string) *GaugeVec {
	return newVector[*Gauge, gnew[Gauge]](opts, GaugeType, labelNames, gnew[Gauge]{})
}

func (v *Gauge) Inc()          { v.Add(1) }
func (v *Gauge) Dec()          { v.Add(-1) }
func (v *Gauge) Sub(d float64) { v.Add(-d) }
func (v *Gauge) Add(d float64) {
	defer v.mu.Unlock()
	v.mu.Lock()

	v.v += d
}

func (v *Gauge) Set(x float64) {
	defer v.mu.Unlock()
	v.mu.Lock()

	v.v = x
}

func NewCounter(opts CounterOpts) *CounterMetric {
	return &CounterMetric{
		d: opts.desc(CounterType),
	}
}

func NewCounterVec(opts CounterOpts, labelNames []string) *CounterVec {
	return newVector[*Counter, gnew[Counter]](opts, CounterType, labelNames, gnew[Counter]{})
}

func (v *Counter) Inc() { v.Add(1) }
func (v *Counter) Add(d float64) {
	if d < 0 {
		panic(d)
	}

	defer v.mu.Unlock()
	v.mu.Lock()

	v.v += d
}

func newVector[T writerMetric, A alloc[T]](opts Opts, typ Type, labelNames []string, a A) *Vector[T, A] {
	return &Vector[T, A]{
		d:      opts.desc(typ),
		labels: labelNames,
		hash:   map[uintptr][]int{},

		a: a,
	}
}

func (v *Vector[T, A]) WithLabelValues(labels ...string) T {
	if len(v.labels) != len(labels) {
		panic(len(v.labels) - len(labels))
	}

	defer v.mu.Unlock()
	v.mu.Lock()

	h := labelsHash(labels)
	vs := v.hash[h]

	for i := 0; i < len(vs); i++ {
		el := v.list[vs[i]]

		if labelsEqual(el.labels, labels) {
			return el.v
		}
	}

	n := vecElem[T]{
		labels: labels,
		v:      v.a.new(),
	}

	idx := len(v.list)

	v.list = append(v.list, n)
	v.hash[h] = append(v.hash[h], idx)

	return n.v
}

func labelsEqual(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}

	return true
}

func labelsHash(ls []string) uintptr {
	var h uintptr

	for _, l := range ls {
		h = strhash(l, h)
	}

	return h
}

type gnew[T any] struct{}

func (gnew[T]) new() *T {
	return new(T)
}
