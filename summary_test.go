package prometheus

import "testing"

func TestSummary(tb *testing.T) {
	var c *Summary

	c.Observe(1)

	c = NewSummary(Opts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "summary",

		Help: "simple summary",

		Labels: []Label{{"counter", "label"}},
	}, SummaryOpts{
		Quantiles: []float64{0.9, 0.95, 1},
	})

	c.Observe(0.1)
	c.Observe(0.2)
	c.Observe(0.3)
	c.Observe(0.4)
	c.Observe(0.5)
	c.Observe(0.6)

	var cv *SummaryVec

	cv.WithLabelValues("first", "second").Observe(1)

	cv = NewSummaryVec(Opts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "summary_vector",

		Help: "simple summary vec",

		Labels: []Label{{"counter", "label"}},
	}, SummaryOpts{
		Quantiles: []float64{0.9, 0.95, 1},
	}, []string{"first"})

	cv.WithLabelValues("a").Observe(0.5)
	cv.WithLabelValues("a").Observe(0.6)
	cv.WithLabelValues("b").Observe(0.7)

	dumpCollectors(tb, c, cv)
}
