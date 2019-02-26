# Prometheusx: helpful functions for working with prometheus

Prometheus extensions to start the prometheus server.

Right now, there is just enough boilerplate in starting a Prometheus server that
we have a lot of half-baked implementations of "start the Prometheus server".
This package provides a single correct way of doing that.

This package also provides a Prometheus linter that can be used in your own code
as follows:

```go
func TestPrometheusMetrics(t *testing.T) {
  prometheusx.LintMetrics(t)
}
```

This will verify that your prometheus metrics follow best practices for
names. The Prometheus best practices are not all obvious, so using this test
in your code is recommended.
