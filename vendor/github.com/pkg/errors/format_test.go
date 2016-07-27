package errors

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"
)

func TestFormatNew(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{{
		New("error"),
		"%s",
		"error",
	}, {
		New("error"),
		"%v",
		"error",
	}, {
		New("error"),
		"%+v",
		"error\n" +
			"github.com/pkg/errors.TestFormatNew\n" +
			"\t.+/github.com/pkg/errors/format_test.go:25",
	}}

	for _, tt := range tests {
		testFormatRegexp(t, tt.error, tt.format, tt.want)
	}
}

func TestFormatErrorf(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{{
		Errorf("%s", "error"),
		"%s",
		"error",
	}, {
		Errorf("%s", "error"),
		"%v",
		"error",
	}, {
		Errorf("%s", "error"),
		"%+v",
		"error\n" +
			"github.com/pkg/errors.TestFormatErrorf\n" +
			"\t.+/github.com/pkg/errors/format_test.go:51",
	}}

	for _, tt := range tests {
		testFormatRegexp(t, tt.error, tt.format, tt.want)
	}
}

func TestFormatWrap(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{{
		Wrap(New("error"), "error2"),
		"%s",
		"error2: error",
	}, {
		Wrap(New("error"), "error2"),
		"%v",
		"error2: error",
	}, {
		Wrap(New("error"), "error2"),
		"%+v",
		"error\n" +
			"github.com/pkg/errors.TestFormatWrap\n" +
			"\t.+/github.com/pkg/errors/format_test.go:77",
	}, {
		Wrap(io.EOF, "error"),
		"%s",
		"error: EOF",
	}, {
		Wrap(New("error with space"), "context"),
		"%q",
		`"context: error with space"`,
	}}

	for _, tt := range tests {
		testFormatRegexp(t, tt.error, tt.format, tt.want)
	}
}

func TestFormatWrapf(t *testing.T) {
	tests := []struct {
		error
		format string
		want   string
	}{{
		Wrapf(New("error"), "error%d", 2),
		"%s",
		"error2: error",
	}, {
		Wrap(io.EOF, "error"),
		"%v",
		"error: EOF",
	}, {
		Wrap(io.EOF, "error"),
		"%+v",
		"EOF\n" +
			"error\n" +
			"github.com/pkg/errors.TestFormatWrapf\n" +
			"\t.+/github.com/pkg/errors/format_test.go:111",
	}, {
		Wrapf(New("error"), "error%d", 2),
		"%v",
		"error2: error",
	}, {
		Wrapf(New("error"), "error%d", 2),
		"%+v",
		"error\n" +
			"github.com/pkg/errors.TestFormatWrapf\n" +
			"\t.+/github.com/pkg/errors/format_test.go:122",
	}, {
		Wrap(Wrap(io.EOF, "error1"), "error2"),
		"%+v",
		"EOF\n" +
			"error1\n" +
			"github.com/pkg/errors.TestFormatWrapf\n" +
			"\t.+/github.com/pkg/errors/format_test.go:128\n",
	}}

	for _, tt := range tests {
		testFormatRegexp(t, tt.error, tt.format, tt.want)
	}
}

func testFormatRegexp(t *testing.T, arg interface{}, format, want string) {
	got := fmt.Sprintf(format, arg)
	lines := strings.SplitN(got, "\n", -1)
	for i, w := range strings.SplitN(want, "\n", -1) {
		match, err := regexp.MatchString(w, lines[i])
		if err != nil {
			t.Fatal(err)
		}
		if !match {
			t.Errorf("fmt.Sprintf(%q, err): got: %q, want: %q", format, got, want)
		}
	}
}
