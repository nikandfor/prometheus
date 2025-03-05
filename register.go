package prometheus

import (
	"net/http"
	"strings"
	"sync"
)

type (
	Registerer interface {
		Register(c Collector) error
		MustRegister(c ...Collector)
		//	Unregister(c Collector) bool
	}

	Registry struct {
		mu sync.Mutex

		l []Collector
		b BufWriter
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

var DefaultRegistry = &Registry{}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) RegisterAll(c ...Collector) error {
	for _, c := range c {
		err := r.Register(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Registry) Register(c Collector) error {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.l = append(r.l, c)

	return nil
}

func (r *Registry) MustRegister(c ...Collector) {
	for _, c := range c {
		err := r.Register(c)
		if err != nil {
			panic(err)
		}
	}
}

// func (r *Registry) Unregister(c Collector) bool { return false }

func (r *Registry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer r.mu.Unlock()
	r.mu.Lock()

	r.b.Reset()

	for _, c := range r.l {
		save := r.b.Len()

		err := c.Collect(&r.b)
		if err != nil {
			r.b.Truncate(save)
		}
	}

	_, err := r.b.WriteTo(w)
	_ = err
}

func (opts *Opts) FQDN() string {
	return join(opts.Namespace, opts.Subsystem, opts.Name)
}

func (opts *Opts) desc(typ Type) Desc {
	ln, lv := splitLabels(opts.Labels)

	return Desc{
		fqdn: opts.FQDN(),
		typ:  typ,
		help: opts.Help,

		labelsNames:  ln,
		labelsValues: lv,
	}
}

func (d *Desc) WriteHeader(w Writer) error {
	return w.Header(d.fqdn, d.help, d.typ, d.labelsNames, d.labelsValues)
}

func (v *Gauge) Collect(w Writer) error {
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

func (v *Counter) Collect(w Writer) error {
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

func (v *Vector[T]) Collect(w Writer) error {
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
