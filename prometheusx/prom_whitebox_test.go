package prometheusx

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/m-lab/go/rtx"
)

func readMetricNumber(t *testing.T, srv *http.Server, label string) float64 {
	metricReader, err := http.Get("http://" + srv.Addr + "/metrics")
	rtx.Must(err, "Could not GET metrics")
	metricBytes, err := ioutil.ReadAll(metricReader.Body)
	rtx.Must(err, "Could not read metrics")

	sawMetric := false
	for _, line := range strings.Split(string(metricBytes), "\n") {
		if strings.HasPrefix(line, "git_short_commit") {
			log.Println("LINE:", line)
			if !strings.Contains(line, "{commit=\""+label+"\"}") {
				continue
			}
			sawMetric = true
			chunks := strings.Split(line, " ")
			if len(chunks) < 2 {
				t.Errorf("Wrong number of pieces in %q", line)
			}
			n, err := strconv.ParseFloat(chunks[len(chunks)-1], 64)
			rtx.Must(err, "Could not convert %q data to an int", line)
			return n
		}
	}
	if !sawMetric {
		t.Errorf("git_short_commit was not found in prometheus output")
	}
	return 0
}

func TestSetCommit(t *testing.T) {
	srv := MustStartPrometheus(":9999")
	defer srv.Shutdown(context.Background())

	tests := []struct {
		commit string
		number float64
	}{
		{"624b332", 103068466},
		{"0", 0},
		{"a", 10},
		{"A", 10},
		{"bad value", 0},
		{"624b332dirty", 0},
	}
	for _, tt := range tests {
		t.Run(tt.commit, func(t *testing.T) {
			setCommitNumber(tt.commit)
			if commit := readMetricNumber(t, srv, tt.commit); commit != tt.number {
				t.Errorf("Bad commit number: %f", commit)
			}
		})
	}
}
