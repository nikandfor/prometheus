package prometheus

import (
	"fmt"
	"io"
)

type (
	Collector interface {
		Collect(w Writer) error
	}

	MetricCollector interface {
		CollectMetric(w Writer, labels Labels) error
	}

	Writer interface {
		Header(fqname, help string, typ Type, constLabels Labels) error
		Write(v float64, suffix string, labels Labels) error
	}

	BufWriter struct {
		// cached last header

		fqname string

		ls   []Label // const labels + buffer for metric labels
		cnst int

		// tmp buffer
		b []byte
	}
)

func NewBufWriter(ls []Label) *BufWriter {
	return &BufWriter{
		ls:   append([]Label{}, ls...),
		cnst: len(ls),
	}
}

func (w *BufWriter) Reset() {
	w.fqname = ""
	w.b = w.b[:0]
}

func (w *BufWriter) Header(fqname, help string, typ Type, ls Labels) error {
	w.fqname = fqname
	w.ls = append(w.ls[:w.cnst], ls...)

	w.b = fmt.Appendf(w.b, "# HELP %s %v\n# TYPE %s %v\n", fqname, help, fqname, typeNames[typ])

	return nil
}

func (w *BufWriter) Write(v float64, suffix string, ls Labels) error {
	w.b = append(w.b, []byte(w.fqname)...)

	if suffix != "" {
		w.b = append(w.b, '_')
		w.b = append(w.b, []byte(suffix)...)
	}

	lab := func(l Label, comma bool) {
		if comma {
			w.b = append(w.b, ',')
		}

		w.b = append(w.b, []byte(l.Name)...)

		w.b = append(w.b, '=', '"')
		w.b = append(w.b, []byte(l.Value)...)
		w.b = append(w.b, '"')
	}

	if len(w.ls)+len(ls) != 0 {
		w.b = append(w.b, '{')
	}

	for i, l := range w.ls {
		lab(l, i != 0)
	}
	for i, l := range ls {
		lab(l, len(w.ls)+i != 0)
	}

	if len(w.ls)+len(ls) != 0 {
		w.b = append(w.b, '}')
	}

	w.b = fmt.Appendf(w.b, " %g\n", v)

	return nil
}

func (w *BufWriter) Bytes() []byte {
	return w.b
}

func (w *BufWriter) Len() int {
	return len(w.b)
}

func (w *BufWriter) Truncate(l int) {
	w.b = w.b[:l]
}

func (w *BufWriter) WriteTo(wr io.Writer) (int64, error) {
	n, err := wr.Write(w.b)
	return int64(n), err
}
