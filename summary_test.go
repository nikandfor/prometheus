package prometheus

import "testing"

func TestSummary(tb *testing.T) {
	var c *Summary

	c.Observe(1)

	c = NewSummary(SummaryOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "summary",

		Help: "simple summary",

		ConstLabels: Labels{{"counter", "label"}},
	})

	c.Observe(0.0)
	c.Observe(0.1)
	c.Observe(0.2)
	c.Observe(0.3)
	c.Observe(0.4)
	c.Observe(0.5)
	c.Observe(0.6)
	c.Observe(0.7)
	c.Observe(0.8)
	c.Observe(0.9)
	c.Observe(1)

	var cv *SummaryVec

	cv.WithLabelValues("first", "second").Observe(1)

	cv = NewSummaryVec(SummaryOpts{
		Namespace: "prometheus",
		Subsystem: "tests",
		Name:      "summary_vector",

		Help: "simple summary vec",

		ConstLabels: Labels{{"counter", "label"}},
	}, []string{"first"})

	cv.WithLabelValues("a").Observe(0.5)
	cv.WithLabelValues("a").Observe(0.6)
	cv.WithLabelValues("b").Observe(0.7)

	dumpCollectors(tb, c, cv)
}
