package errorx

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestSuppress(t *testing.T) {
	derivedFromUnexpectedEOF := fmt.Errorf("Derived from %w", io.ErrUnexpectedEOF)
	looksLikeUnexpectedEOFButIsnt := fmt.Errorf("Contains %q but is not derived from it", io.ErrUnexpectedEOF)
	type args struct {
		incomingError    error
		errorsToSuppress []error
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "Filter ErrServerClosed",
			args: args{http.ErrServerClosed, []error{http.ErrServerClosed}},
			want: nil,
		},
		{
			name: "Filter WithNil",
			args: args{http.ErrServerClosed, []error{nil}},
			want: http.ErrServerClosed,
		},
		{
			name: "Filter WithNothing",
			args: args{http.ErrServerClosed, []error{nil}},
			want: http.ErrServerClosed,
		},
		{
			name: "Filter Nil",
			args: args{nil, []error{http.ErrServerClosed}},
			want: nil,
		},
		{
			name: "Filter NilNil",
			args: args{nil, []error{nil}},
			want: nil,
		},
		{
			name: "Filter NilNothing",
			args: args{nil, []error{}},
			want: nil,
		},
		{
			name: "Filter Many Errors",
			args: args{http.ErrServerClosed, []error{io.ErrClosedPipe, io.ErrUnexpectedEOF, io.ErrNoProgress}},
			want: http.ErrServerClosed,
		},
		{
			name: "Filter Derived Errors",
			args: args{derivedFromUnexpectedEOF, []error{io.ErrClosedPipe, io.ErrUnexpectedEOF, io.ErrNoProgress}},
			want: nil,
		},
		{
			name: "Filter Derived Errors But Not Strings",
			args: args{looksLikeUnexpectedEOFButIsnt, []error{io.ErrClosedPipe, io.ErrUnexpectedEOF, io.ErrNoProgress}},
			want: looksLikeUnexpectedEOFButIsnt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Suppress(tt.args.incomingError, tt.args.errorsToSuppress...)
			if got != tt.want {
				t.Errorf("Suppress() = %v, want %v", got, tt.want)
			}
		})
	}
}
