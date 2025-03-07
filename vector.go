package prometheus

import "sync"

type (
	Vector[Metric MetricCollector] struct {
		d Desc

		labels []string // variable labels names

		a Allocator[Metric]

		mu sync.Mutex

		list []vecElem[Metric]
		hash map[uintptr][]int
	}

	vecElem[Metric MetricCollector] struct {
		labels Labels // values

		m Metric
	}

	Allocator[T any] interface {
		New() T
	}

	BasicAllocator[T any] struct{}
)

func NewVector[Metric MetricCollector](o Opts, typ Type, labelNames []string, a Allocator[Metric]) *Vector[Metric] {
	return &Vector[Metric]{
		d: NewDescFromOpts(o, typ),

		labels: append([]string{}, labelNames...),

		a:    a,
		hash: make(map[uintptr][]int),
	}
}

func (m *Vector[Metric]) Collect(w Writer) error {
	err := m.d.CollectHeader(w)
	if err != nil {
		return err
	}

	defer m.mu.Unlock()
	m.mu.Lock()

	for i := range m.list {
		err = m.list[i].m.CollectMetric(w, m.list[i].labels)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Vector[Metric]) WithLabelValues(labelValues ...string) Metric {
	if m == nil {
		var zero Metric
		return zero
	}

	h := valuesHash(labelValues)

	defer m.mu.Unlock()
	m.mu.Lock()

	if len(labelValues) != len(m.labels) {
		panic(len(labelValues) - len(m.labels))
	}

	vs := m.hash[h]

	for i := 0; i < len(vs); i++ {
		el := m.list[vs[i]]

		if labelsToValuesEqual(el.labels, labelValues) {
			return el.m
		}
	}

	ls := make(Labels, len(m.labels))

	for i, n := range m.labels {
		ls[i] = Label{Name: n, Value: labelValues[i]}
	}

	n := vecElem[Metric]{
		labels: ls,
		m:      m.a.New(),
	}

	idx := len(m.list)

	m.list = append(m.list, n)
	m.hash[h] = append(m.hash[h], idx)

	return n.m
}

func labelsValuesEqual(x, y Labels) bool {
	if len(x) != len(x) {
		return false
	}

	for i := range x {
		if x[i].Value != y[i].Value {
			return false
		}
	}

	return true
}

func labelsToValuesEqual(x Labels, y []string) bool {
	if len(x) != len(x) {
		return false
	}

	for i := range x {
		if x[i].Value != y[i] {
			return false
		}
	}

	return true
}

func valuesEqual(x, y []string) bool {
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

func labelsValuesHash(ls []Label) uintptr {
	var h uintptr

	for _, l := range ls {
		h = strhash(l.Value, h)
	}

	return h
}

func valuesHash(ls []string) uintptr {
	var h uintptr

	for _, l := range ls {
		h = strhash(l, h)
	}

	return h
}

func (a BasicAllocator[T]) New() *T { return new(T) }
