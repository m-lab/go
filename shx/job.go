// Package shx provides shell-like operations for Go.
//
// A Job represents one or more operations. A single-operation Job may represent
// running a command (Exec, System), or reading or writing a file (ReadFile,
// WriteFile), or a user defined operation (Func). A multiple-operation Job runs
// several single operation jobs in a sequence (Script) or pipeline (Pipe).
// Taken together, these primitive types allow the composition of more and more
// complex operations.
//
// Users control how a Job runs using State. State controls the Job input and
// output, as well as its working directory and environment.
//
// Users may produce a human-readable representation of a complex Job in a
// shell-like syntax using the Description. Because some operations have no
// shell equivalent, the result is only representative.
//
// Examples are provided for all primitive Job types: Exec, System, Func, Pipe,
// Script. Additional convenience Jobs make creating more complex operations a
// little easier. Advanced users may create their own Job types for more
// flexibility.
package shx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Description is used to produce a representation of a Job. Custom Job types
// should use the Description interface to represent their behavior in a
// helpful, human-readable form. After collecting a description, serialize using
// the String() method.
type Description struct {
	// Depth is used to control line prefix indentation in complex Jobs.
	Depth int
	desc  bytes.Buffer
	line  int
	seps  []string
	idxs  []int
}

// Append adds a new command to the output buffer. Typically, the command is
// formatted as a new line at the end of the current buffer. If StartSequence
// was called before calling Append, then the command is formatted as a
// continuation of the current line.
func (d *Description) Append(cmd string) {
	l := len(d.idxs)
	if l > 0 {
		d.idxs[l-1]++
		if d.idxs[l-1] > 1 {
			// After the first cmd, separate others with a separator.
			d.desc.WriteString(d.seps[l-1] + cmd)
			return
		}
		d.desc.WriteString(cmd)
		return
	}
	d.line++
	d.desc.WriteString(fmt.Sprintf("%2d: %s%s\n", d.line, prefix(d.Depth), cmd))
}

// StartSequence begins formatting a multi-part expression on a single line,
// such as a list, pipeline, or similar expression. StartSequence begins with
// "start" and subsequent calls to Append add commands to the end of the current
// line, separating sequential commands with "sep". StartSequence returns a
// function that ends the line and restores the default behavior of Append.
func (d *Description) StartSequence(start, sep string) (endlist func(end string)) {
	d.seps = append(d.seps, sep)
	d.idxs = append(d.idxs, 0)
	l := len(d.idxs)
	if l == 1 {
		d.line++
		d.desc.WriteString(fmt.Sprintf("%2d: %s", d.line, prefix(d.Depth)))
	}
	if l > 1 {
		// For deeper nesting, use the prior separator prior to current start.
		d.desc.WriteString(d.seps[l-2])
	}
	d.desc.WriteString(start)
	endlist = func(end string) {
		l := len(d.idxs)
		// Verify that some commands were printed before adding extra newline.
		d.desc.WriteString(end)
		if l == 1 {
			d.desc.WriteString("\n")
		}
		d.seps = d.seps[:len(d.seps)-1]
		d.idxs = d.idxs[:len(d.idxs)-1]
	}
	return endlist
}

// String serializes a description produced by running Job.Describe(). Calling
// String resets the Description buffer.
func (d *Description) String() string {
	s := d.desc.String()
	d.desc.Reset()
	return s
}

// State is a Job configuration. Callers provide the first initial State, and
// as a Job executes it creates new State instances derived from the original,
// e.g. for Pipes and subcommands.
type State struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Dir    string
	Env    []string
}

// New creates a State instance based on the current process state, using
// os.Stdin, os.Stdout, and os.Stderr, as well as the current working directory
// and environment.
func New() *State {
	d, _ := os.Getwd()
	s := &State{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Dir:    d,
		Env:    os.Environ(),
	}
	return s
}

func prefix(d int) string {
	v := ""
	for i := 0; i < d; i++ {
		v = v + "  "
	}
	return v
}

// SetDir assigns the Dir value and returns the previous value.
func (s *State) SetDir(dir string) string {
	prev := s.Dir
	s.Dir = dir
	return prev
}

// Path produces a path relative to the State's current directory. If arguments
// represent an absolute path, then that is used. If multiple arguments are
// provided, they're joined using filepath.Join.
func (s *State) Path(path ...string) string {
	if len(path) == 0 {
		return s.Dir
	}
	if filepath.IsAbs(path[0]) {
		return filepath.Join(path...)
	}
	return filepath.Join(append([]string{s.Dir}, path...)...)
}

// SetEnv assigns the named variable to the given value in the State
// environment. If the named variable is already defined it is overwritten.
func (s *State) SetEnv(name, value string) {
	prefix := name + "="
	// Find and overwrite an existing value.
	for i, kv := range s.Env {
		if strings.HasPrefix(kv, prefix) {
			s.Env[i] = prefix + value
			return
		}
	}
	// Or, add the new value to the s.Env.
	s.Env = append(s.Env, prefix+value)
}

// GetEnv reads the named variable from the State environment. If name is not
// found, an empty value is returned. An undefined variable and a variable set
// to the empty value are indistinguishable.
func (s *State) GetEnv(name string) string {
	prefix := name + "="
	for _, kv := range s.Env {
		if strings.HasPrefix(kv, prefix) {
			return strings.TrimPrefix(kv, prefix)
		}
	}
	// name not found.
	return ""
}

// Job is the interface for an operation. A Job controls how an operation is run
// and represented.
type Job interface {
	// Describe produces a readable representation of the Job operation. After
	// calling Describe, use Description.String() to report the result.
	Describe(d *Description)

	// Run executes the Job using the given State. A Job should terminate when
	// the given context is cancelled.
	Run(ctx context.Context, s *State) error
}

// Exec creates a Job to execute the given command with the given arguments.
func Exec(cmd string, args ...string) *ExecJob {
	return &ExecJob{
		name: cmd,
		args: args,
	}
}

// System is an Exec job that interprets the given command using "/bin/sh".
func System(cmd string) *ExecJob {
	return &ExecJob{
		name: "/bin/sh",
		args: []string{"-c", cmd},
	}
}

// ExecJob implements the Job interface for basic process execution.
type ExecJob struct {
	name string
	args []string
}

// Run executes the command.
func (f *ExecJob) Run(ctx context.Context, s *State) error {
	cmd := exec.CommandContext(ctx, f.name, f.args...)
	cmd.Dir = s.Dir
	cmd.Env = s.Env
	cmd.Stdin = s.Stdin
	cmd.Stdout = s.Stdout
	cmd.Stderr = s.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s %s", err, f.name, strings.Join(f.args, " "))
	}
	return nil
}

// Describe generates a description for this command.
func (f *ExecJob) Describe(d *Description) {
	args := ""
	if len(f.args) > 0 {
		args = " " + strings.Join(f.args, " ")
	}
	d.Append(f.name + args)
}
