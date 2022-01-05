package base

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"text/template"
)

const (
	tpl = `|{{"Metric" | printf "%-20s"}}|{{"Times"|printf "%10s"}}|{{"Avg"|printf "%13s"}}|{{"Min"|printf "%13s"}}|{{"Max"|printf "%13s"}}|
{{- range $x := . -}}
{{if (gt (call $x.Times $x) 0)}}
|{{call $x.Name $x | printf "%-20s" -}}|{{call $x.Times $x | printf "%10d" }}|{{call $x.Avg $x | printf "%10s" }}|{{call $x.Min $x | printf "%10s" }}|{{call $x.Max $x | printf "%10s|"}}
{{- end}}
{{- end}}`
)

type Metric struct {
	name   string
	num    uint64
	elapse uint64
	min    uint64
	max    uint64
	Name   func(*Metric) string
	Times  func(*Metric) uint64
	Avg    func(*Metric) string
	Min    func(*Metric) string
	Max    func(*Metric) string
}

func SemanticTime(f float64) string {
	if (f - 1000000.0) > 0.00000001 {
		// ms
		return fmt.Sprintf("%10.2f ms", f/1000000.0)
	} else if (f - 1000.0) > 0.00000001 {
		// us
		return fmt.Sprintf("%10.2f us", f/1000.0)
	} else {
		return fmt.Sprintf("%10.2f ns", f)
	}
}

func Name(m *Metric) string {
	return m.name
}

func Times(m *Metric) uint64 {
	return m.num
}

func Avg(m *Metric) string {
	return SemanticTime(float64(m.elapse) / float64(m.num))
}

func Min(m *Metric) string {
	return SemanticTime(float64(m.min))
}

func Max(m *Metric) string {
	return SemanticTime(float64(m.max))
}

func NewMetric() *Metric {
	return &Metric{
		Name:  Name,
		Times: Times,
		Avg:   Avg,
		Min:   Min,
		Max:   Max,
	}
}

func (m *Metric) Add(data uint64) {
	atomic.AddUint64(&m.num, 1)
	atomic.AddUint64(&m.elapse, data)
	if m.min == 0 {
		m.min = data
		m.max = data
	} else if data < m.min {
		m.min = data
	} else if data > m.max {
		m.max = data
	}
}

func (m *Metric) Reset() {
	if m.num == 0 {
		return
	}

	m.num, m.elapse, m.min, m.max = 0, 0, 0, 0
}

// 只支持预先串行分配
type MetricContainer struct {
	title string
	ms    []*Metric
	tpl   *template.Template
}

func NewMetricContainer(title string) (*MetricContainer, error) {
	tpl, err := template.New(title).Parse(tpl)
	if err != nil {
		return nil, err
	}

	return &MetricContainer{
		title: title,
		ms:    make([]*Metric, 0, 3),
		tpl:   tpl,
	}, nil
}

func (mc *MetricContainer) Alloc(name string) (metricAlias int) {
	M := NewMetric()
	M.name = name
	mc.ms = append(mc.ms, M)
	return len(mc.ms) - 1
}

func (mc *MetricContainer) AddMetric(alias int, data uint64) {
	mc.ms[alias].Add(data)
}

func (mc *MetricContainer) Reset() {
	for i := 0; i < len(mc.ms); i++ {
		mc.ms[i].Reset()
	}
}

func (mc *MetricContainer) Count() (string, error) {
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString(fmt.Sprintf("\n|%s\n", mc.title))
	err := mc.tpl.Execute(buffer, mc.ms)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}
