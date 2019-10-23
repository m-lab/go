Functions which extend the capabilities of the `flag` package.

This package includes:

* `flagx.ArgsFromEnv` allows flags to be passed from the command-line or as an
  environment variable.

* `flagx.FileBytes` is a new flag type. It automatically reads the content of
  the given file as a `[]byte`, handling any error during flag parsing and
  simplifying application logic.

* `flagx.StringArray` is a new flag type that handles appending to `[]string`

Usage of any of the above is like:

```Go
package main

import (
	"flag"
	"fmt"

	"github.com/m-lab/go/flagx"
)

var (
	flagArray flagx.StringArray
)

func main() {
	flag.Var(&flagArray, "array", "append to string array")
	flag.Parse()
	fmt.Printf("%+v\n", flagArray)
	// your code here
}
```
