package prometheus

import "strings"

type (
	Register struct {
		//
	}

	Collector interface {
		Collect(w Writer) error
	}

	Writer interface {
		Header(fqdn, help string, typ Type, labelNames, labelValues []string) error
		Metric(suffix string, v float64, labelNames, labelValues []string) error
	}

	Desc struct {
		fqdn string
		help string
		typ  Type

		labelsNames  []string
		labelsValues []string
	}
)

func (opts *Opts) desc(typ Type) Desc {
	ln, lv := splitLabels(opts.Labels)

	return Desc{
		fqdn: join(opts.Namespace, opts.Subsystem, opts.Name),
		typ:  typ,
		help: opts.Help,

		labelsNames:  ln,
		labelsValues: lv,
	}
}

func (d *Desc) WriteHeader(w Writer) error {
	return w.Header(d.fqdn, d.help, d.typ, d.labelsNames, d.labelsValues)
}

func (v *GaugeMetric) Collect(w Writer) error {
	err := v.d.WriteHeader(w)
	if err != nil {
		return err
	}

	return v.writeMetric(w, nil, nil)
}

func (v *Gauge) writeMetric(w Writer, ln, lv []string) error {
	defer v.mu.Unlock()
	v.mu.Lock()

	err := w.Metric("", v.v, ln, lv)
	if err != nil {
		return err
	}

	return nil
}

func (v *CounterMetric) Collect(w Writer) error {
	err := v.d.WriteHeader(w)
	if err != nil {
		return err
	}

	return v.writeMetric(w, nil, nil)
}

func (v *Counter) writeMetric(w Writer, ln, lv []string) error {
	defer v.mu.Unlock()
	v.mu.Lock()

	err := w.Metric("", v.v, ln, lv)
	if err != nil {
		return err
	}

	return nil
}

func (v *Vector[T, A]) Collect(w Writer) error {
	err := v.d.WriteHeader(w)
	if err != nil {
		return err
	}

	defer v.mu.Unlock()
	v.mu.Lock()

	for _, m := range v.list {
		err = m.v.writeMetric(w, v.labels, m.labels)
		if err != nil {
			return err
		}
	}

	return nil
}

func splitLabels(ls []Label) (ln, lv []string) {
	ln = make([]string, len(ls))
	lv = make([]string, len(ls))

	for i, l := range ls {
		ln[i] = l.Name
		lv[i] = l.Value
	}

	return ln, lv
}

func join(s ...string) string {
	var b strings.Builder

	for _, s := range s {
		if s == "" {
			continue
		}

		if b.Len() != 0 {
			_ = b.WriteByte('_')
		}

		_, _ = b.WriteString(strings.TrimFunc(s, func(r rune) bool { return r == '_' }))
	}

	return b.String()
}
