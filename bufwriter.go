package prometheus

import (
	"fmt"
	"io"
)

type (
	BufWriter struct {
		prefix string
		fqdn   string

		ln, lv []string
		static int

		b []byte
	}
)

func NewBufWriter(ls []Label) *BufWriter {
	ln, lv := splitLabels(ls)

	return &BufWriter{
		ln:     append([]string{}, ln...),
		lv:     append([]string{}, lv...),
		static: len(ln),
	}
}

func (w *BufWriter) Reset() {
	w.fqdn = ""
	w.b = w.b[:0]
}

func (w *BufWriter) Bytes() []byte {
	return w.b
}

func (w *BufWriter) WriteTo(wr io.Writer) (int64, error) {
	n, err := wr.Write(w.b)
	return int64(n), err
}

func (w *BufWriter) Header(fqdn, help string, typ Type, ln, lv []string) error {
	if len(ln) != len(lv) {
		panic(len(ln) - len(lv))
	}

	w.fqdn = fqdn
	w.ln = append(w.ln[:w.static], ln...)
	w.lv = append(w.lv[:w.static], lv...)

	c := ""

	if w.prefix != "" {
		c = "_"
	}

	w.b = fmt.Appendf(w.b, "# HELP %v%v%v %v\n# TYPE %v%v%v %v\n", w.prefix, c, fqdn, help, w.prefix, c, fqdn, typeNames[typ])

	return nil
}

func (w *BufWriter) Metric(suffix string, v float64, ln, lv []string) error {
	if len(ln) != len(lv) {
		panic(len(ln) - len(lv))
	}

	if w.prefix != "" {
		w.b = append(w.b, []byte(w.prefix)...)
		w.b = append(w.b, '_')
	}

	w.b = append(w.b, []byte(w.fqdn)...)

	if suffix != "" {
		w.b = append(w.b, '_')
		w.b = append(w.b, []byte(suffix)...)
	}

	lab := func(n, v string, comma bool) {
		if comma {
			w.b = append(w.b, ',')
		}

		w.b = append(w.b, []byte(n)...)

		w.b = append(w.b, '=', '"')
		w.b = append(w.b, []byte(v)...)
		w.b = append(w.b, '"')
	}

	if len(w.ln)+len(ln) != 0 {
		w.b = append(w.b, '{')
	}

	for i := range w.ln {
		lab(w.ln[i], w.lv[i], i != 0)
	}
	for i := range ln {
		lab(ln[i], lv[i], len(w.ln)+i != 0)
	}

	if len(w.ln)+len(ln) != 0 {
		w.b = append(w.b, '}')
	}

	w.b = fmt.Appendf(w.b, " %g\n", v)

	return nil
}
