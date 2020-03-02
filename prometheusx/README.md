# Prometheusx: helpful functions for working with prometheus

Prometheus extensions to expose and lint Prometheus metrics.

## `MustServeMetrics` and `--prometheusx.listen-address`

Right now, there is just enough boilerplate in starting an http server that
exposes Prometheus metrics that we have a lot of half-baked implementations
of "expose the Prometheus metrics". This package provides a single correct
way of doing that via `MustServeMetrics` and provides a unified command-line
flag `--prometheusx.listen-address` used by that function.  The older function
`MustStartPrometheus` is deprecated and should not be used in new code.

## Exporting commit hashes in metrics

The variable `GitShortCommit` is intended to hold the truncated commit id for
a git commit. It should be equal to the first column of the output of `git
log --oneline`. This string is interpreted by `init()` as a base-16 number
and the Prometheus metric `git_short_commit` is set to the resulting
numerical value and the commit hash is also used as the label for that
metric. It is recommended that the string be set as part of the build/link
process, as follows:

```bash
  go build -ldflags "-X prometheusx.GitShortCommit=$(git log -1 --format=%h)$(git diff --quiet || echo dirty)" ./...
```

This metric should be useful when determining whether code on various
systems is running the same version, which should, among other things, help
detect failed rollouts or extended periods in which test deployments occur
but never a production deployment.

### FAQ: Why don't we also put in the tag?

The tag is a property of the repo, not the code. You add tags without
changing the code, and if two images are built from the same commit then they
should have the same binaries bit-for-bit. Putting the tag in the binary is
bad for the same reason putting the build time in the binary is bad: It
forever prevents builds from being hermetic, and
[hermetic builds](https://landing.google.com/sre/sre-book/chapters/release-engineering/)
are a good north star at which to aim.

## Linting metrics

The subpackage `promtest` also provides a Prometheus linter that can be used in your own code
as follows:

```go
func TestPrometheusMetrics(t *testing.T) {
  promtest.LintMetrics(t)
}
```

This will verify that your prometheus metrics follow best practices for
names. The Prometheus best practices are not all obvious, so using this test
in your code is recommended. Note that all metrics whcih have labels will
need to be incremented at least once in order to appear in the output and be
linted. Caveat coder.
