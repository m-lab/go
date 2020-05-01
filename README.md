# go
| branch | travis-ci | coveralls | docs | report card |
|--------|-----------|-----------|------|-------------|
| master | [![Travis Build Status](https://travis-ci.org/m-lab/go.svg?branch=master)](https://travis-ci.org/m-lab/go) | [![Coverage Status](https://coveralls.io/repos/m-lab/go/badge.svg?branch=master)](https://coveralls.io/github/m-lab/go?branch=master) | [![GoDoc](https://godoc.org/github.com/m-lab/go?status.svg)](https://godoc.org/github.com/m-lab/go) | [![Go Report Card](https://goreportcard.com/badge/github.com/m-lab/go)](https://goreportcard.com/report/github.com/m-lab/go)

General purpose libraries / APIs for use in mlab code.

## General guidance
Packages in this repo should be:
+ Useful across multiple other packages
+ Non-trivial, either in lines of code or in semantic complexity.
Small simple things should likely just be defined where they are used.
+ Fairly carefully designed.  Probably should review design with other
engineers before putting in too much effort.
+ Well tested and well documented.  Test and documentation standards
should be even higher than for most code repositories.

Note that packages here are intended to be used in *other* repositories.
This means that it will be somewhat disruptive to change APIs in these
packages, as API changing PRs will break other repos, and require additional
PRs to fix those repositories.

## tags

Please never tag this repo as v1.0 or above.  Each library within this repo
exists independently, and the commitments required by Go module best practices
can never be satisfied by this repo as a whole.  According to Go best practices,
no whole-repo promises are made for tags of the form `v0.X.Y`, so we will
restrict ourselves to version tags that start with a zero.

Please mark packages in development as *alpha* or *beta*.  Use of these packages
should be discouraged in other repositories while those packages are under
development.

## packages
[link to go docs](https://pkg.go.dev/github.com/m-lab/go)
