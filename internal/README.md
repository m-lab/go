This directory contains:

1. internal code copied from "cloud.google.com/go/internal/..." which is used to support adapting bigquery code to create a fake for bigquery.Uploader.
2. the code adapted from "cloud.google.com/go/bigquery", in internal/frombigquery

The files should be direct copies of the code in cloud.google.com, aside from internal cross references.

Ideally, this code should be kept up to date to reflect changes in "cloud.google.com/go/internal/...".
