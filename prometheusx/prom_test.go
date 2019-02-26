package prometheusx_test

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/m-lab/go/prometheusx"
)

func TestMustStartPrometheus(t *testing.T) {
	srv := prometheusx.MustStartPrometheus(":9999")
	defer srv.Shutdown(context.Background())
	if srv.Addr != ":9999" {
		t.Error("We should get back any non-empty address we pass in")
	}
}

func TestMustStartPrometheusOnEmptyAddr(t *testing.T) {
	srv := prometheusx.MustStartPrometheus(":0")
	defer srv.Shutdown(context.Background())
	if srv.Addr == ":0" {
		t.Error("We should never get back an empty address")
	}
}

func TestLintMetricsEmpty(t *testing.T) {
	// No metrics.
	prometheusx.LintMetrics(t)
	if !prometheusx.LintMetrics(nil) {
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
	prometheusx.LintMetrics(t)
	if !prometheusx.LintMetrics(nil) {
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
	prometheusx.LintMetrics(subT)
	if !subT.Failed() {
		t.Errorf("On bad metrics the test should fail")
	}
	if prometheusx.LintMetrics(nil) {
		t.Error("Failed to lint error on one bad metric")
	}
}
