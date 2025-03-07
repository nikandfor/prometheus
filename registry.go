package prometheus

type (
	Registry struct {
		ms []Collector
	}
)

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Collect(w Writer) error {
	for _, c := range r.ms {
		err := c.Collect(w)
		if err != nil {
			return err
		}
	}

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

func (r *Registry) Register(c Collector) error {
	r.ms = append(r.ms, c)

	return nil
}
