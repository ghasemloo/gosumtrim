package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	root       = "./" // module root
	inModFile  = root + "testdata/go.mod.in"
	inSumFile  = root + "testdata/go.sum.in"
	outSumFile = root + "testdata/go.sum.out"
)

func TestTrim(t *testing.T) {
	mod := mustOpen(t, inModFile)
	sum := mustOpen(t, inSumFile)
	out := mustOpen(t, outSumFile)
	want, err := ioutil.ReadAll(out)
	if err != nil {
		t.Fatalf("ioutil.ReadAll(out) failed: %v", err)
	}

	b := bytes.NewBuffer(nil)
	if err := trim(mod, sum, b); err != nil {
		t.Fatalf("trime() failed: %v", err)
	}
	got := b.Bytes()

	t.Logf("got:\n%v", string(got))
	t.Logf("want:\n%v", string(want))

	if diff := cmp.Diff(string(got), string(want)); diff != "" {
		t.Fatalf("trime() returned:\n%s", diff)
	}
}

func mustOpen(t *testing.T, fn string) io.Reader {
	t.Helper()
	r, err := os.Open(fn)
	if err != nil {
		t.Fatalf("os.Open(%q) failed: %v", fn, err)
	}
	return r
}
