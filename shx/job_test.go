package shx_test

import (
	"bytes"
	"context"
	"log"
	"os"
	"testing"

	. "github.com/m-lab/go/shx"
)

func init() {
	log.SetFlags(log.LUTC | log.Llongfile)
}

func TestDescription(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		cmds  []string
		want  string
	}{
		{
			name:  "success-script",
			lines: []string{"env", "pwd"},
			want:  " 1: env\n 2: pwd\n 3: \n",
		},
		{
			name: "success-pipe",
			cmds: []string{"env", "cat"},
			want: " 1: env | cat\n",
		},
		{
			name:  "success-script-pipe",
			lines: []string{"env", "pwd"},
			cmds:  []string{"env", "cat"},
			want:  " 1: env\n 2: pwd\n 3: env | cat\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Description{}
			for _, line := range tt.lines {
				d.Append(line)
			}
			endlist := d.StartSequence("", " | ")
			for _, cmd := range tt.cmds {
				d.Append(cmd)
			}
			endlist("")
			v := d.String()
			if v != tt.want {
				t.Errorf("Description: wrong result; got %q, want %q", v, tt.want)
			}
		})
	}
}

func TestExec(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "success",
			cmd:  "/bin/echo",
			args: []string{"a", "b"},
			want: "a b\n",
		},
		{
			name:    "error-no-such-command",
			cmd:     "/not-a-dir/not-a-real-command",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Exec(tt.cmd, tt.args...)
			ctx := context.Background()
			b := bytes.NewBuffer(nil)
			s := &State{
				Stdout: b,
			}
			err := job.Run(ctx, s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exec() = %v, want %t", err, tt.wantErr)
			}
			if b.String() != tt.want {
				t.Errorf("Exec() = got %v, want %v", b.String(), tt.want)
			}
		})
	}
}

func TestState(t *testing.T) {
	t.Run("SetState", func(t *testing.T) {
		s := New()
		origDir := s.Dir
		if p := s.SetDir("/"); p != origDir {
			t.Errorf("SetDir() wrong previous value; got %q, want %q", p, origDir)
		}
		s.SetEnv("FOO", "BAR")
		if p := s.GetEnv("FOO"); p != "BAR" {
			t.Errorf("SetEnv() found wrong value; got %q, want %q", p, "BAR")
		}
		// Set the same variable with a new value.
		s.SetEnv("FOO", "BAR2")
		if p := s.GetEnv("FOO"); p != "BAR2" {
			t.Errorf("SetEnv() found wrong value; got %q, want %q", p, "BAR2")
		}
		if p := s.GetEnv("NOTFOUND"); p != "" {
			t.Errorf("GetEnv() found value; got %q, want %q", p, "")
		}
		if p := s.Path(); p != "/" {
			t.Errorf("Path() wrong value; got %q, want %q", p, "/")
		}
		if p := s.Path("/"); p != "/" {
			t.Errorf("Path() wrong value; got %q, want %q", p, "/")
		}
		if p := s.Path("relative"); p != "/relative" {
			t.Errorf("Path() wrong value; got %q, want %q", p, "/relative")
		}
		if p := s.Path("relative", "path"); p != "/relative/path" {
			t.Errorf("Path() wrong value; got %q, want %q", p, "/relative/path")
		}
	})
}

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		job  Job
		want string
	}{
		{
			name: "exec-name-only",
			job:  Exec("ls"),
			want: " 1: ls\n",
		},
		{
			name: "exec-name-with-args",
			job:  Exec("ls", "-l"),
			want: " 1: ls -l\n",
		},
		{
			name: "system-with-args",
			job:  System("ls -l"),
			want: " 1: /bin/sh -c ls -l\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Description{}
			tt.job.Describe(d)
			val := d.String()

			if val != tt.want {
				t.Errorf("Job.Describe() unexpected; got = %q, want %q", val, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		job     Job
		want    string
		wantDir string
		wantEnv string
		wantErr bool
	}{
		{
			name: "exec-echo",
			job:  Exec("echo", "ok"),
			want: "ok\n",
		},
		{
			name:    "exec-error",
			job:     Exec("/this-command-does-not-exist", "ok"),
			wantErr: true,
		},
		{
			name:    "system-error",
			job:     System("exit 1"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bytes.Buffer{}
			s := &State{
				Stdout: b,
			}
			ctx := context.Background()
			err := tt.job.Run(ctx, s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Job.Run() unexpected error; got = %v, wantErr %t", err, tt.wantErr)
			}
			if s.Dir != tt.wantDir {
				t.Errorf("Job.Run() unexpected Dir; got = %q, want %q", s.Dir, tt.wantDir)
			}
			val := b.String()
			if val != tt.want {
				t.Errorf("Job.Run() unexpected output; got = %q, want %q", val, tt.want)
			}
			if s.GetEnv("key") != tt.wantEnv {
				t.Errorf("Job.Run() unexpected output; got = %q, want %q", val, tt.wantEnv)
			}
		})
	}
}

func ExampleExecJob_Run() {
	ex := Exec("echo", "a", "b")
	s := &State{
		Stdout: os.Stdout,
	}
	err := ex.Run(context.Background(), s)
	if err != nil {
		panic(err)
	}
	// Output: a b
}

func ExampleSystem() {
	sys := System("echo a b")
	s := &State{
		Stdout: os.Stdout,
	}
	err := sys.Run(context.Background(), s)
	if err != nil {
		panic(err)
	}
	// Output: a b
}
