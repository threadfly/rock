package base

import (
	"testing"
)

type Log interface {
	Fatalf(format string, args ...interface{})
}

func NewMC(log Log, name string) *MetricContainer {
	mc, err := NewMetricContainer(name)
	if err != nil {
		log.Fatalf("new metric container, %s", err)
	}
	return mc
}

func TestMetric(t *testing.T) {
	mc := NewMC(t, "test")

	HELLO := mc.Alloc("hello")
	WORLD := mc.Alloc("world")
	BIG := mc.Alloc("big")

	mc.AddMetric(HELLO, 350)
	mc.AddMetric(HELLO, 400)
	mc.AddMetric(HELLO, 500)
	mc.AddMetric(WORLD, 1000)
	mc.AddMetric(HELLO, 98394)
	mc.AddMetric(WORLD, 2434)
	mc.AddMetric(BIG, 2434)
	ctt, err := mc.Count()
	if err != nil {
		t.Errorf("metric container count, %s", err)
	}

	t.Log(ctt)
}

func BenchmarkMetric(b *testing.B) {
	mc := NewMC(b, "test")
	BENCH := mc.Alloc("bench")
	b.ResetTimer()
	b.StartTimer()
	for i := uint64(0); i < uint64(b.N); i++ {
		mc.AddMetric(BENCH, i)
	}
}
