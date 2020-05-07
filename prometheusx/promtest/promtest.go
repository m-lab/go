// Package promtest provides a utility function for linting Prometheus metrics.
package promtest

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/m-lab/go/prometheusx"
	"github.com/m-lab/go/rtx"
	"github.com/prometheus/client_golang/prometheus/testutil/promlint"
)

// LintMetrics will ensure that the names of the passed-in Promethus metrics
// follow all best practices. If the passed-in testing.T is nil, then all lint
// errors are just log messages. If a real testing.T is passed in, then lint
// errors cause test failures.
func LintMetrics(t *testing.T) (passed bool) {
	srv := prometheusx.MustStartPrometheus(":0")
	defer srv.Shutdown(context.Background())

	metricReader, err := http.Get("http://" + srv.Addr + "/metrics")
	rtx.Must(err, "Could not GET metrics")
	metricBytes, err := ioutil.ReadAll(metricReader.Body)
	rtx.Must(err, "Could not read metrics")

	metricsLinter := promlint.New(bytes.NewBuffer(metricBytes))
	problems, err := metricsLinter.Lint()
	rtx.Must(err, "Could not lint metrics")

	passed = true
	for _, p := range problems {
		passed = false
		msg := fmt.Sprintf("Bad metric %v: %v", p.Metric, p.Text)
		if t == nil {
			log.Println(msg)
		} else {
			t.Error(msg)
		}
	}
	return passed
}
