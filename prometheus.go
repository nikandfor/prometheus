package prometheus

import "strings"

type (
	Opts struct {
		Namespace string
		Subsystem string
		Name      string

		Help string

		ConstLabels Labels
	}

	Desc struct {
		fqname string
		help   string
		typ    Type
		labels Labels
	}

	Type int

	Labels []Label

	Label struct {
		Name  string
		Value string
	}
)

const (
	CounterType Type = iota
	GaugeType
	SummaryType
	UntypedType
	HistogramType
	GaugeHistogramType
)

func (t Type) String() string { return typeNames[t] }

var typeNames = []string{"gauge", "counter", "summary", "untyped"}

func NewDesc(fqname, help string, typ Type, constLabels Labels) Desc {
	return Desc{
		fqname: fqname,
		help:   help,
		typ:    typ,
		labels: append(Labels{}, constLabels...),
	}
}

func NewDescFromOpts(o Opts, typ Type) Desc {
	return NewDesc(BuildFQName(o.Namespace, o.Subsystem, o.Name), o.Help, typ, o.ConstLabels)
}

func (d *Desc) CollectHeader(w Writer) error {
	return w.Header(d.fqname, d.help, d.typ, d.labels)
}

func BuildFQName(namespace, subsystem, name string) string {
	return join(namespace, subsystem, name)
}

func join(ss ...string) string {
	switch len(ss) {
	case 0:
		return ""
	case 1:
		return ss[0]
	}

	var b strings.Builder

	for _, s := range ss {
		s = strings.Trim(s, "_")
		if s == "" {
			continue
		}

		if b.Len() != 0 {
			_ = b.WriteByte('_')
		}

		_, _ = b.WriteString(s)
	}

	return b.String()
}
