package gometrics

import (
	"runtime"
	"slices"
	"sync"
	"time"

	"nikand.dev/go/prometheus"
)

type (
	GoMetrics struct {
		mu sync.Mutex

		infoLs prometheus.Labels
		bufls  prometheus.Labels
	}
)

func (m *GoMetrics) init() {
	if m.infoLs == nil {
		m.infoLs = prometheus.Labels{{Name: "version", Value: runtime.Version()}}
	}
}

func (m *GoMetrics) Collect(w prometheus.Writer) error {
	m.init()

	err := m.basic(w, "go_info", 1, prometheus.GaugeType, m.infoLs, "Information about the Go environment.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_gc_gogc_percent", float64(1), prometheus.GaugeType, nil, "Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_gc_gomemlimit_bytes", float64(1), prometheus.GaugeType, nil, "Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_goroutines", float64(runtime.NumGoroutine()), prometheus.GaugeType, nil, "Number of goroutines that currently exist.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_sched_gomaxprocs_threads", float64(runtime.GOMAXPROCS(0)), prometheus.GaugeType, nil, "The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads")
	if err != nil {
		return err
	}

	err = m.memstats(w)
	if err != nil {
		return err
	}

	return nil
}

func (m *GoMetrics) basic(w prometheus.Writer, name string, v float64, typ prometheus.Type, ls prometheus.Labels, help string) error {
	err := w.Header(name, help, typ, nil)
	if err != nil {
		return err
	}

	err = w.Write(v, "", ls)
	if err != nil {
		return err
	}

	return nil
}

func (m *GoMetrics) memstats(w prometheus.Writer) error {
	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	slices.Sort(ms.PauseNs[:])
	l := len(ms.PauseNs)

	err := w.Header("go_gc_duration_seconds", "A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.", prometheus.SummaryType, nil)
	if err != nil {
		return err
	}

	m.bufls = append(m.bufls[:0], prometheus.Label{Name: "quantile", Value: "0"})

	err = w.Write(time.Duration(ms.PauseNs[0]).Seconds(), "", m.bufls)
	if err != nil {
		return err
	}

	m.bufls[0].Value = "0.25"

	err = w.Write(time.Duration(ms.PauseNs[l/4]).Seconds(), "", m.bufls)
	if err != nil {
		return err
	}

	m.bufls[0].Value = "0.5"

	err = w.Write(time.Duration(ms.PauseNs[l/2]).Seconds(), "", m.bufls)
	if err != nil {
		return err
	}

	m.bufls[0].Value = "0.75"

	err = w.Write(time.Duration(ms.PauseNs[l*3/4]).Seconds(), "", m.bufls)
	if err != nil {
		return err
	}

	m.bufls[0].Value = "1"

	err = w.Write(time.Duration(ms.PauseNs[l-1]).Seconds(), "", m.bufls)
	if err != nil {
		return err
	}

	err = w.Write(time.Duration(ms.PauseTotalNs).Seconds(), "sum", nil)
	if err != nil {
		return err
	}

	err = w.Write(float64(ms.NumGC), "count", nil)
	if err != nil {
		return err
	}

	err = m.basic(w, "go_memstats_alloc_bytes", float64(ms.Alloc), prometheus.GaugeType, nil, "Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_memstats_alloc_bytes_total", float64(ms.TotalAlloc), prometheus.CounterType, nil, "Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_memstats_heap_objects", float64(ms.Mallocs-ms.Frees), prometheus.GaugeType, nil, "Number of currently allocated objects. Equals to /gc/heap/objects:objects.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_memstats_mallocs_total", float64(ms.Mallocs), prometheus.CounterType, nil, "Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.")
	if err != nil {
		return err
	}

	err = m.basic(w, "go_memstats_next_gc_bytes", float64(ms.NextGC), prometheus.GaugeType, nil, "Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.")
	if err != nil {
		return err
	}

	return nil
}
