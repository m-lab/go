package logx

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/m-lab/go/flagx"
)

var (
	// LogxLevel is an enum used to select the log level.
	LogxLevel = flagx.Enum{
		Options: []string{"warn", "info", "debug"},
	}
)

// Loggers for supported log levels.
var (
	DefaultFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	Warn         = log.New(ioutil.Discard, "WARN : ", DefaultFlags)
	Info         = log.New(ioutil.Discard, "INFO : ", DefaultFlags)
	Debug        = log.New(ioutil.Discard, "DEBUG: ", DefaultFlags)
)

func init() {
	flag.Var(&LogxLevel, "logx.level", "Enable logging at this level and higher")
}

// Loggers returns the debug, info, and warnign loggers for local convenience variables.
func Loggers() (*log.Logger, *log.Logger, *log.Logger) {
	return Debug, Info, Warn
}

// Setup updates the package loggers based on the LogxLevel value.
func Setup() error {
	switch LogxLevel.Value {
	case "debug":
		Debug.SetOutput(os.Stderr)
		fallthrough
	case "info":
		Info.SetOutput(os.Stderr)
		fallthrough
	case "warn":
		Warn.SetOutput(os.Stderr)
	case "":
		// Ignore the empty value.
		return nil
	default:
		// Report an error for unknown values.
		return fmt.Errorf("Unsupported value: %q", LogxLevel.Value)
	}
	return nil
}
