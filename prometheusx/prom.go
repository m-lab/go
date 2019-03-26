// Package prometheusx provides a canonical way to expose Prometheus metrics
// and provides a utility function for linting those metrics.
package prometheusx

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/pprof"
	"strconv"
	"testing"

	"github.com/m-lab/go/httpx"
	"github.com/m-lab/go/rtx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/prometheus/util/promlint"
)

var (
	// GitShortCommit holds the truncated commit id for a git commit. It should be
	// equal to the first column of the output of `git log --oneline`. This string
	// is interpreted by init() as a base-16 number and the Prometheus metric
	// git_short_commit is set to the resulting numerical value. It is recommended
	// that the string be set as part of the build/link process, as follows:
	//
	//   go build -ldflags "-X prometheusx.GitShortCommit=`git log HEAD~1..HEAD --format=tformat:%h`" ./...
	//
	// This metric should be useful when determining whether code on various
	// systems is running the same version, which should, among other things, help
	// detect failed rollouts, or extended periods in which test deployments occur
	// but never a production deployment.
	GitShortCommit       = "No commit specified"
	gitShortCommitMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "git_short_commit",
		Help: "The git commit interpreted as a number.",
	})
)

func setCommitNumber(commit string) {
	number, err := strconv.ParseInt(commit, 16, 64)
	if err == nil {
		gitShortCommitMetric.Set(float64(number))
	} else {
		gitShortCommitMetric.Set(0)
	}
}

func init() {
	setCommitNumber(GitShortCommit)
}

// MustStartPrometheus starts an http server which exposes local metrics to
// Prometheus.  If the passed-in address is ":0" then a random open port will
// be chosen and the .Addr element of the returned server will be udpated to
// reflect the actual port.
func MustStartPrometheus(addr string) *http.Server {
	// Prometheus with some extras.
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/metrics", promhttp.Handler())

	// Start up the http server.
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	rtx.Must(httpx.ListenAndServeAsync(server), "Could not start metric server")

	// Return the server
	return server
}

// LintMetrics will ensure that the names of the passed-in Promethus metrics
// follow all best practices. If the passed-in testing.T is nil, then all lint
// errors are just log messages. If a real testing.T is passed in, then lint
// errors cause test failures.
func LintMetrics(t *testing.T) (passed bool) {
	srv := MustStartPrometheus(":0")
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
