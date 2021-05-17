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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

func (s *State) copy() *State {
	c := &State{
		Stdin:  s.Stdin,
		Stdout: s.Stdout,
		Stderr: s.Stderr,
		Dir:    s.Dir,
	}
	// Make independent copy of environment.
	c.Env = append(c.Env, s.Env...)
	return c
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

// Func creates a new FuncJob that runs the given function. Job functions should
// honor the context to support cancelation. The given name is used to describe
// this function.
func Func(name string, job func(ctx context.Context, s *State) error) *FuncJob {
	return &FuncJob{
		Job: job,
		Desc: func(d *Description) {
			d.Append(name)
		},
	}
}

// FuncJob is a generic Job type that allows creating new operations without
// creating a totally new type. When created directly, both Job and Desc fields
// must be defined.
type FuncJob struct {
	Job  func(ctx context.Context, s *State) error
	Desc func(d *Description)
}

// Run executes the job function.
func (f *FuncJob) Run(ctx context.Context, s *State) error {
	return f.Job(ctx, s)
}

// Describe generates a description for this custom function.
func (f *FuncJob) Describe(d *Description) {
	f.Desc(d)
}

// Chdir creates Job that changes the State Dir to the given directory at
// runtime. This does not alter the process working directory. Chdir is helpful
// in Script() Jobs.
func Chdir(dir string) Job {
	return &FuncJob{
		Job: func(ctx context.Context, s *State) error {
			s.Dir = s.Path(dir)
			return nil
		},
		Desc: func(d *Description) {
			d.Append(fmt.Sprintf("cd %s", dir))
		},
	}
}

// Println writes the given message to the State Stdout and expands variable
// references from the running State environment. Println supports the same
// variable syntax as os.Expand, e.g. $NAME or ${NAME}.
func Println(message string) Job {
	return &FuncJob{
		Job: func(ctx context.Context, s *State) error {
			message = os.Expand(message, s.GetEnv)
			_, err := s.Stdout.Write([]byte(message + "\n"))
			return err
		},
		Desc: func(d *Description) {
			d.Append(fmt.Sprintf("echo %q", message))
		},
	}
}

// SetEnv creates a Job to assign the given name=value in the running State Env.
// SetEnv is helpful in Script() Jobs.
func SetEnv(name string, value string) Job {
	return &FuncJob{
		Job: func(ctx context.Context, s *State) error {
			s.SetEnv(name, value)
			return nil
		},
		Desc: func(d *Description) {
			d.Append(fmt.Sprintf("export %s=%q", name, value))
		},
	}
}

// SetEnvFromJob creates a new Job that sets the given name in Env to the result
// written to stdout by running the given Job. Errors from the given Job are
// returned.
func SetEnvFromJob(name string, job Job) Job {
	return &FuncJob{
		Job: func(ctx context.Context, s *State) error {
			b := &bytes.Buffer{}
			s2 := &State{
				Stdout: b,
				Env:    append([]string(nil), s.Env...),
			}
			err := job.Run(ctx, s2)
			if err != nil {
				return err
			}
			s.SetEnv(name, strings.TrimSpace(b.String()))
			return nil
		},
		Desc: func(d *Description) {
			close := d.StartSequence(fmt.Sprintf("export %s=$(", name), "")
			job.Describe(d)
			close(")")
		},
	}
}

// IfFileMissing creates a Job that runs the given job if the named file does
// not exist.
func IfFileMissing(file string, job Job) Job {
	return &FuncJob{
		Desc: func(d *Description) {
			d.Append(fmt.Sprintf("if [[ ! -f %s ]] ; then", file))
			d.Depth++
			job.Describe(d)
			d.Depth--
			d.Append("fi")
		},
		Job: func(ctx context.Context, s *State) error {
			_, err := os.Stat(s.Path(file))
			if err != nil {
				return job.Run(ctx, s)
			}
			// This is not an error, we simply don't run the job.
			return nil
		},
	}
}

// IfVarEmpty creates a Job that runs the given job if the named variable is
// empty.
func IfVarEmpty(key string, job Job) Job {
	return &FuncJob{
		Desc: func(d *Description) {
			d.Append(fmt.Sprintf("if [[ -z ${%s} ]] ; then", key))
			d.Depth++
			job.Describe(d)
			d.Depth--
			d.Append("fi")
		},
		Job: func(ctx context.Context, s *State) error {
			if s.GetEnv(key) == "" {
				return job.Run(ctx, s)
			}
			// This is not an error, we simply don't run the job.
			return nil
		},
	}
}

// Script creates a Job that executes the given Job parameters in sequence. If
// any Job returns an error, execution stops.
func Script(t ...Job) *ScriptJob {
	return &ScriptJob{
		Jobs: t,
	}
}

// ErrScriptError is a base Script error.
var ErrScriptError = errors.New("script execution error")

// ScriptJob implements the Job interface for running an ordered sequence of Jobs.
type ScriptJob struct {
	Jobs []Job
}

// Run sequentially executes every Job in the script. Any Job error stops
// execution and generates an error describing the command that failed.
func (c *ScriptJob) Run(ctx context.Context, s *State) error {
	z := s.copy()
	for i := range c.Jobs {
		err := c.Jobs[i].Run(ctx, z)
		// Only generate description when the error is NOT a script error.
		if err != nil && !errors.Is(err, ErrScriptError) {
			d := &Description{}
			c.Describe(d)
			str := d.String()
			return fmt.Errorf("%w:\n%s - %s", ErrScriptError, str, err.Error())
		}
		// All other errors.
		if err != nil {
			return err
		}
	}
	return nil
}

// Describe generates a description for all jobs in the script.
func (c *ScriptJob) Describe(d *Description) {
	d.Append("(")
	d.Depth++
	for i := range c.Jobs {
		c.Jobs[i].Describe(d)
	}
	d.Depth--
	d.Append(")")
}

// Pipe creates a Job that executes the given Jobs as a "shell pipeline",
// passing the output of the first to the input of the next, and so on.
// If any Job returns an error, the first error is returned.
func Pipe(t ...Job) *PipeJob {
	return &PipeJob{
		Jobs: t,
	}
}

// PipeJob implements the Job interface for running multiple Jobs in a
// pipeline.
type PipeJob struct {
	Jobs []Job
}

// Run executes every Job in the pipeline. The stdout from the first command is
// passed to the stdin to the next command. The stderr for all commands is
// inherited from the given State. If any Job returns an error, the first error
// is returned for the entire PipeJob.
func (c *PipeJob) Run(ctx context.Context, z *State) error {
	e := c.Jobs
	p := nPipes(z.Stdin, z.Stdout, len(e))
	s := make([]*State, len(e))
	for i := range e {
		s[i] = &State{
			Stdin:  p[i].R,
			Stdout: p[i].W,
			Stderr: z.Stderr,
			Dir:    z.Dir,
			Env:    z.Env,
		}
	}
	// Create channel for all pipe job return values.
	done := make(chan error, len(e))

	// Create a wait group to block on all Jobs returning.
	wg := sync.WaitGroup{}
	defer wg.Wait()

	// Context cancellation will execute before waiting on wait group.
	ctx2, cancel2 := context.WithCancel(ctx)
	defer cancel2()

	// Run all jobs in reverse order, end of pipe to beginning.
	for i := len(e) - 1; i >= 0; i-- {
		wg.Add(1)
		go func(n, i int, e Job, s *State) {
			err := e.Run(ctx2, s)
			// Send possible errors to outer loop.
			done <- err
			wg.Done()
			if i != 0 {
				closeReader(s.Stdin)
			}
			if i != n-1 {
				closeWriter(s.Stdout)
			}
		}(len(e), i, e[i], s[i])
	}

	// Wait for goroutines to return or context cancellation.
	for range e {
		var err error
		select {
		case err = <-done:
		case <-ctx.Done():
			// Continue collecting errors after context cancellation.
			err = <-done
		}
		// Return first error. Deferred wait group will block until all Jobs return.
		if err != nil {
			return err
		}
	}
	return nil
}

// Describe generates a description for all jobs in the pipeline.
func (c *PipeJob) Describe(d *Description) {
	endlist := d.StartSequence("", " | ")
	defer endlist("")
	for i := range c.Jobs {
		c.Jobs[i].Describe(d)
	}
}

func closeWriter(w io.Writer) error {
	c, ok := w.(io.WriteCloser)
	if ok {
		return c.Close()
	}
	// Not a write closer, so cannot be closed.
	return nil
}

func closeReader(r io.Reader) error {
	c, ok := r.(io.ReadCloser)
	if ok {
		return c.Close()
	}
	// Not a read closer, so cannot be closed.
	return nil
}

type rw struct {
	R io.Reader
	W io.Writer
}

func nPipes(r io.Reader, w io.Writer, n int) []rw {
	var p []rw
	for i := 0; i < n-1; i++ {
		rp, wp := io.Pipe()
		p = append(p, rw{R: r, W: wp})
		r = rp
	}
	p = append(p, rw{R: r, W: w})
	return p
}
