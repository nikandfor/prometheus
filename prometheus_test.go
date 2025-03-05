package prometheus

import "testing"

func TestGauge(tb *testing.T) {
	var c *Gauge

	c.Inc()

	c = NewGauge(GaugeOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "gauge",

		Help: "simple gauge",

		Labels: []Label{{"counter", "label"}},
	})

	c.Add(4)
	c.Inc()

	var cv *GaugeVec

	cv.WithLabelValues("first", "second").Inc()

	cv = NewGaugeVec(GaugeOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "gauge_vector",

		Help: "simple vectorized gauge",

		Labels: []Label{{"vector", "label"}},
	}, []string{"first", "second"})

	cv.WithLabelValues("1", "a").Inc()
	cv.WithLabelValues("1", "b").Add(2)

	dumpCollectors(tb, c, cv)
}

func TestCounter(tb *testing.T) {
	var c *Counter

	c.Inc()

	c = NewCounter(CounterOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "counter",

		Help: "simple counter",

		Labels: []Label{{"counter", "label"}},
	})

	c.Add(4)
	c.Inc()

	var cv *CounterVec

	cv.WithLabelValues("first", "second").Inc()

	cv = NewCounterVec(CounterOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "counter_vector",

		Help: "simple vectorized counter",

		Labels: []Label{{"vector", "label"}},
	}, []string{"first", "second"})

	cv.WithLabelValues("1", "a").Inc()
	cv.WithLabelValues("1", "b").Add(2)

	dumpCollectors(tb, c, cv)
}

func dumpCollectors(tb testing.TB, cc ...Collector) {
	w := NewBufWriter([]Label{{"buf", "label"}})

	for _, c := range cc {
		err := c.Collect(w)
		assertNoError(tb, err)
	}

	tb.Logf("dump:\n%s", w.Bytes())
}

func assertNoError(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Errorf("error: %v", err)
	}
}
