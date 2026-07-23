package main

import (
	"bytes"
	"flag"
	"go/format"
	"os"
	"path/filepath"
	"testing"
)

func NoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

var update bool

func init() {
	flag.BoolVar(&update, "update", update, "Update golden files")
}

func TestBetterIfErr(t *testing.T) {
	tests, err := os.ReadDir("./testdata")
	NoError(t, err)
	for _, test := range tests {
		if test.IsDir() {
			t.Run(test.Name(), func(t *testing.T) {
				source := filepath.Join("./testdata", test.Name(), "source.go")
				sourceAbs, err := filepath.Abs(source)
				NoError(t, err)

				sourceBytes, err := os.ReadFile(sourceAbs)
				NoError(t, err)

				pos := bytes.Index(sourceBytes, []byte("/*error*/"))
				if pos == -1 {
					t.Errorf("Expected /*error*/ not found")
				}

				resultStr, err := generateErrReturn(sourceAbs, sourceBytes, pos)
				NoError(t, err)

				resultBytes, err := format.Source([]byte(resultStr))
				NoError(t, err)
				resultBytes = bytes.TrimSuffix(resultBytes, []byte("\n"))

				want := filepath.Join("./testdata", test.Name(), "result.golden")
				if update {
					os.WriteFile(want, resultBytes, 0644)
					return
				}
				wantBytes, err := os.ReadFile(want)
				NoError(t, err)
				wantBytes = bytes.TrimSuffix(wantBytes, []byte("\n"))

				if !bytes.Equal(resultBytes, wantBytes) {
					t.Errorf("Expected \n%s\n, got \n%s\n", string(wantBytes), string(resultBytes))
				}
			})
		}
	}
}
