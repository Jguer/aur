package main

import (
	"bytes"
	"io"
	"testing"

	"github.com/Jguer/aur"
	"github.com/stretchr/testify/assert"
)

func Test_printSearch(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer

	a := &aur.Pkg{
		Name:        "test",
		Version:     "1.0.0.",
		NumVotes:    20,
		Popularity:  4.0,
		Description: "Test description",
	}

	testWriter := io.Writer(&b)

	printSearch(a, testWriter)

	assert.Equal(t, "- \x1b[1mtest\x1b[0m 1.0.0. (20 4.00)\n\tTest description\n", b.String())
}

func Test_printInfo(t *testing.T) {
	t.Setenv("TZ", "UTC")

	a := &aur.Pkg{
		Name:           "test",
		Version:        "1.0.0.",
		NumVotes:       20,
		Popularity:     4.0,
		Description:    "Test description",
		LastModified:   0,
		FirstSubmitted: 0,
	}

	tests := []struct {
		name    string
		verbose bool
		wantW   string
	}{
		{
			name:    "verbose",
			verbose: true,
			wantW:   "",
		},
		{
			name:    "not verbose",
			verbose: false,
			wantW:   "\x1b[1mName            : \x1b[0mtest\n\x1b[1mVersion         : \x1b[0m1.0.0.\n\x1b[1mDescription     : \x1b[0mTest description\n\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var b bytes.Buffer
			testWriter := io.Writer(&b)

			printInfo(a, testWriter, "https://aur.archlinux.org", tt.verbose)
			if !tt.verbose {
				assert.Equal(t, tt.wantW, b.String())
			}
		})
	}
}
