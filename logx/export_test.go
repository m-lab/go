package logx

import "os"

func BadPipeForTest() {
	pipe = func() (r *os.File, w *os.File, err error) {
		return nil, nil, os.ErrNotExist
	}
}

func RestorePipeForTest() {
	pipe = os.Pipe
}
