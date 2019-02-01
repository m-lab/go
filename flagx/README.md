Functions which extend the capabilities of the `flag` package.

This package includes:

* `flagx.ArgsFromEnv` allows flags to be passed from the command-line or as an
  environment variable.

* `flagx.FileBytes` is a new flag type. It automatically reads the content of
  the given file as a `[]byte`, handling any error during flag parsing and
  simplifying application logic.
