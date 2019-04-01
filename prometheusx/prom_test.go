package prometheusx_test

import (
	"context"
	"testing"

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
