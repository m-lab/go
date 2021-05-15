package shx_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
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

func TestFunc(t *testing.T) {
	count := 0
	tests := []struct {
		name string
		job  func(ctx context.Context, s *State) error
	}{
		{
			name: "success",
			job:  func(ctx context.Context, s *State) error { count++; return nil },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Func(tt.name, tt.job)
			ctx := context.Background()
			s := &State{
				Stdout: os.Stdout,
			}
			err := f.Run(ctx, s)
			if err != nil {
				t.Errorf("Func() failed; got %v, want nil", err)
			}
		})
	}
	if count != 1 {
		t.Errorf("Func() count incorrect; got %d, want 1", count)
	}
}

func TestScript(t *testing.T) {
	tmpdir := t.TempDir()

	tests := []struct {
		name    string
		t       []Job
		want    string
		wantErr bool
	}{
		{
			name: "success",
			t: []Job{
				Chdir(tmpdir),
				System("pwd"),
			},
			want: tmpdir + "\n",
		},
		{
			name: "stop-after-error",
			t: []Job{
				// Force an error.
				System("exit 1"),
				Func("test-failure", func(ctx context.Context, s *State) error {
					t.Fatalf("script should not continue executing after error.")
					return nil
				}),
			},
			wantErr: true,
		},
		{
			name: "stop-after-deep-error",
			t: []Job{
				// Force an error within a sub-Script.
				Script(
					System("exit 1"),
				),
				Func("test-failure", func(ctx context.Context, s *State) error {
					t.Fatalf("script should not continue executing after error.")
					return nil
				}),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			b := bytes.NewBuffer(nil)
			s := &State{
				Stdout: b,
			}
			sc := Script(tt.t...)
			err := sc.Run(ctx, s)
			if (err != nil) && !tt.wantErr {
				t.Fatalf("failed to run test: %s", err)
			}
			if b.String() != tt.want {
				t.Errorf("Script() wrong pwd output; got %s, want %s", b.String(), tt.want)
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
			name: "func-simple",
			job:  Func("simple", func(ctx context.Context, s *State) error { return nil }),
			want: " 1: simple\n",
		},
		{
			name: "func-custom",
			job: &FuncJob{
				Job: func(ctx context.Context, s *State) error { return nil },
				Desc: func(d *Description) {
					d.Append("custom")
				},
			},
			want: " 1: custom\n",
		},
		{
			name: "script",
			job:  Script(Exec("echo", "ok")),
			want: " 1: (\n 2:   echo ok\n 3: )\n",
		},
		{
			name: "func-chdir",
			job:  Chdir("otherdir"),
			want: " 1: cd otherdir\n",
		},
		{
			name: "func-setenv",
			job:  SetEnv("key", "value"),
			want: ` 1: export key="value"` + "\n",
		},
		{
			name: "func-setenvfromjob",
			job:  SetEnvFromJob("key", Exec("echo", "ok")),
			want: " 1: export key=$(echo ok)\n",
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
			name: "func-custom",
			job: &FuncJob{
				Job: func(ctx context.Context, s *State) error {
					s.Stdout.Write([]byte("output"))
					return nil
				},
				Desc: func(d *Description) {},
			},
			want: "output",
		},
		{
			name: "script",
			job:  Script(Exec("echo", "ok")),
			want: "ok\n",
		},
		{
			name:    "script-error",
			job:     Script(System("exit 1")),
			wantErr: true,
		},
		{
			name:    "script-deep-error",
			job:     Script(Script(System("exit 1"))),
			wantErr: true,
		},
		{
			name:    "func-chdir",
			job:     Chdir("otherdir"),
			wantDir: "otherdir",
		},
		{
			name:    "func-setenv",
			job:     SetEnv("key", "value"),
			wantEnv: "value",
		},
		{
			name: "func-setenv-overwrite",
			job: Func("reset", func(ctx context.Context, s *State) error {
				s.SetEnv("key", "original")
				s.SetEnv("key", "final")
				return nil
			}),
			wantEnv: "final",
		},
		{
			name:    "func-setenvfromjob",
			job:     SetEnvFromJob("key", Exec("echo", "value")),
			wantEnv: "value",
		},
		{
			name:    "func-setenvfromjob-error",
			job:     SetEnvFromJob("key", System("exit 1")),
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

func ExampleFuncJob_Run() {
	f := Func("example", func(ctx context.Context, s *State) error {
		b, err := ioutil.ReadAll(s.Stdin)
		if err != nil {
			return err
		}
		_, err = s.Stdout.Write([]byte(base64.URLEncoding.EncodeToString(b)))
		return err
	})
	s := &State{
		Stdin:  bytes.NewBuffer([]byte(`{"key":"value"}\n`)),
		Stdout: os.Stdout,
	}
	err := f.Run(context.Background(), s)
	if err != nil {
		panic(err)
	}
	// Output: eyJrZXkiOiJ2YWx1ZSJ9XG4=
}

func ExampleScriptJob_Run() {
	sc := Script(
		SetEnv("FOO", "BAR"),
		Exec("env"),
	)
	s := &State{
		Stdout: os.Stdout,
	}
	err := sc.Run(context.Background(), s)
	if err != nil {
		panic(err)
	}
	// Output: FOO=BAR
}

func ExampleScriptJob_Describe() {
	sc := Script(
		SetEnv("FOO", "BAR"),
		Exec("env"),
	)
	d := &Description{}
	sc.Describe(d)
	fmt.Println("\n" + d.String())
	// Output:
	//  1: (
	//  2:   export FOO="BAR"
	//  3:   env
	//  4: )
}

func Example() {
	sc := Script(
		// Set environment in Script State.
		SetEnv("KEY", "ORIGINAL"),
		Script(
			// Overwrite environment in sub-Script.
			SetEnv("KEY", "SUBSCRIPT"),
			Exec("env"),
		),
		// Original Script State environment was not modified by sub-Script.
		Exec("env"),
		// Overwrite environment using command output.
		SetEnvFromJob("KEY", System("basename $( pwd )")),
		Exec("env"),
	)
	s := New()
	s.Env = nil // Clear state environment for example.
	err := sc.Run(context.Background(), s)
	if err != nil {
		panic(err)
	}
	// Output: KEY=SUBSCRIPT
	// KEY=ORIGINAL
	// KEY=shx
}
