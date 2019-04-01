package promtest_test

import (
	"testing"

	"github.com/m-lab/go/prometheusx/promtest"
	"github.com/prometheus/client_golang/prometheus"
)

func TestLintMetricsEmpty(t *testing.T) {
	// No metrics.
	promtest.LintMetrics(t)
	if !promtest.LintMetrics(nil) {
		t.Error("Failed to lint empty metrics")
	}
}

func TestLintMetricsGoodMetric(t *testing.T) {
	// One good metric.
	goodC := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "good_total",
		Help: "Counts good things",
	})
	prometheus.MustRegister(goodC)
	defer prometheus.Unregister(goodC)
	promtest.LintMetrics(t)
	if !promtest.LintMetrics(nil) {
		t.Error("Failed to lint one good metric")
	}
}

func TestLintMetricsBadMetric(t *testing.T) {
	// One bad metric.
	badC := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "bad_name_name_for_a_counter",
	})
	prometheus.MustRegister(badC)
	defer prometheus.Unregister(badC)
	subT := &testing.T{}
	promtest.LintMetrics(subT)
	if !subT.Failed() {
		t.Errorf("On bad metrics the test should fail")
	}
	if promtest.LintMetrics(nil) {
		t.Error("Failed to lint error on one bad metric")
	}
}
